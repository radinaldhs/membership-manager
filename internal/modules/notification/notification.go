package notification

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/cenkalti/backoff/v5"
	"github.com/panjf2000/ants/v2"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

// Message topics
const (
	topicGeneral = "GENERAL"
	topicPromo   = "PROMO"
)

// FCM topics
const (
	fcmTopicGeneral = "authenticated.general"
	fcmTopicPromo   = "authenticated.promo"
)

// Message topic to FCM topic map
var topicToFcmMsgTopic = map[string]string{
	topicGeneral: fcmTopicGeneral,
	topicPromo:   fcmTopicPromo,
}

// Message topic to mobile UI routing
var topicToMobileUiRouting = map[string]string{
	// TODO: Consult with mobile team on this
	// topicGeneral: "/notifications",
	topicPromo: "/promo",
}

// Notification statuses
const (
	notifPending   = "PENDING"
	notifSent      = "SENT"
	notifDelivered = "DELIVERED"
	notifRead      = "READ"
)

type ActionData struct {
	// Use this to tell client which page should be showed when
	// the notification is clicked
	Screen string `json:"screen"`
	// Use this to tell client that the content they are looking
	// for is provided by requesting through this URL. This can
	// be embedded web view, RESTful API endpoint, etc.
	URL string `json:"url"`
}

type Action struct {
	Topic        string        `json:"topic"`
	TargetFilter *TargetFilter `json:"target_filter"`
	Data         *ActionData   `json:"data"`
}

// TODO: Custom validation
type TargetFilter struct {
	MemberType  []int    `json:"member_type"`
	Gender      []string `json:"gender"`
	Age         []int    `json:"age"`
	MemberCodes []string `json:"member_codes"`
}

// TODO: Add validation
type Message struct {
	Title    string `json:"title" validate:"required"`
	Body     string `json:"body" validate:"required"`
	ImageURL string `json:"image_url" validate:"required"`
	Action   Action `json:"action"`
}

type ClientNotificationStatus struct {
	NotificationID    int    `json:"notification_id" validate:"required"`
	FCMNotificationID string `json:"fcm_notification_id" validate:"required"`
	Status            string `json:"status" validate:"required"`
}

type fcmNotificationData map[string]string

func (fcmNotifData fcmNotificationData) setNotificationId(id int) {
	fcmNotifData["notification_id"] = strconv.Itoa(id)
}

func (fcmNotifdata fcmNotificationData) setScreen(screen string) {
	fcmNotifdata["screen"] = screen
}

func (fcmNotifData fcmNotificationData) setUrl(url string) {
	fcmNotifData["url"] = url
}

type NotificationProvider interface {
	SaveFCMToken(ctx context.Context, fcmToken FCMToken) error
	Push(ctx context.Context, msg Message) error
	SetMemberNotificationStatus(ctx context.Context, memberId, memberRegistBranchId int, notifStatus ClientNotificationStatus) error
	GetMemberNotificationList(ctx context.Context, memberId, memberRegistBranchId int, page, pageSize int) ([]MemberNotificationListItemViewModel, error)
	DeleteMemberNotification(ctx context.Context, memberId, memberRegistId, notifId int) error
	Close() error
}

// TODO: Method to return available topics to listen to
type Notification struct {
	cfg          config.Notification
	notifRepo    NotificationRepository
	fcmClient    *messaging.Client
	logger       *logging.Logger
	gPool        *ants.Pool
	isClosed     *atomic.Bool
	retryBackoff backoff.BackOff
}

func NewNotification(cfg config.Notification, notifRepo NotificationRepository, fcmClient *messaging.Client, logger *logging.Logger) (*Notification, error) {
	gPool, err := ants.NewPool(cfg.ConcurrentPushLimit)
	if err != nil {
		return nil, err
	}

	return &Notification{
		cfg:       cfg,
		notifRepo: notifRepo,
		fcmClient: fcmClient,
		logger:    logger,
		gPool:     gPool,
		isClosed:  &atomic.Bool{},
		// TODO: Custom configuration for backoff
		retryBackoff: backoff.NewExponentialBackOff(),
	}, nil
}

func (notif *Notification) SaveFCMToken(ctx context.Context, fcmToken FCMToken) error {
	_, err := notif.fcmClient.SubscribeToTopic(ctx, []string{fcmToken.Token}, fcmTopicGeneral)
	if err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	_, err = notif.fcmClient.SubscribeToTopic(ctx, []string{fcmToken.Token}, fcmTopicPromo)
	if err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	return notif.notifRepo.SaveFCMToken(ctx, fcmToken)
}

// TODO: Not really sure if we really need this
func (notif *Notification) Topics() []string {
	var topics []string
	for _, t := range topicToFcmMsgTopic {
		topics = append(topics, t)
	}

	return topics
}

func (notif *Notification) Push(ctx context.Context, msg Message) error {
	if closed := notif.isClosed.Load(); closed {
		return customerror.ErrServiceUnavailable
	}

	go notif._pushMsg(msg)

	return nil
}

func (notif *Notification) SetMemberNotificationStatus(ctx context.Context, memberId, memberRegistBranchId int, notifStatus ClientNotificationStatus) error {
	// TODO: Validate object
	memberNotifModel := MemberNotificationModel{
		MemberID:                   memberId,
		MemberRegistrationBranchID: memberRegistBranchId,
		NotificationID:             notifStatus.NotificationID,
		FCMMessageID:               &notifStatus.FCMNotificationID,
		Status:                     notifStatus.Status,
	}

	return notif.notifRepo.SaveMemberNotification(ctx, memberNotifModel)
}

func (notif *Notification) GetMemberNotificationList(ctx context.Context, memberId, memberRegistBranchId int, page, pageSize int) ([]MemberNotificationListItemViewModel, error) {
	if pageSize == 0 {
		// TODO: Store this value in an constant
		pageSize = 50
	}

	if page < 0 {
		page = 1
	}

	offset := (page - 1) * pageSize
	list, err := notif.notifRepo.GetMemberNotifications(ctx, memberId, memberRegistBranchId, offset, pageSize)
	if err != nil {
		return nil, err
	}

	if list == nil {
		list = []MemberNotificationListItemViewModel{}
	}

	return list, nil
}

func (notif *Notification) DeleteMemberNotification(ctx context.Context, memberId, memberRegistId, notifId int) error {
	return notif.notifRepo.DeleteMemberNotification(ctx, memberId, memberRegistId, notifId)
}

func (notif *Notification) push(ctx context.Context, fcmMsg *messaging.Message, msgModel NotificationMessageModel, token *FCMToken) {
	var memberNotifModel MemberNotificationModel
	if token != nil {
		memberNotifModel = MemberNotificationModel{
			MemberID:                   token.MemberID,
			MemberRegistrationBranchID: token.MemberRegistrationBranchID,
			NotificationID:             msgModel.ID,
			FCMMessageID:               nil,
			Status:                     notifPending,
		}
	}

	err := notif.notifRepo.SaveMemberNotification(ctx, memberNotifModel)

	if err != nil {
		logMsg, logAttrs := logAttrsFailToSendMsg(msgModel, token, err)
		notif.logger.Error(logMsg, logAttrs...)

		return
	}

	_, err = backoff.Retry(ctx, backoff.Operation[string](func() (string, error) {
		fcmId, err := notif.fcmClient.Send(ctx, fcmMsg)
		if err != nil {
			if messaging.IsUnavailable(err) {
				// TODO: Set retry after from config
				return "", backoff.RetryAfter(5)
			}

			if messaging.IsQuotaExceeded(err) {
				// TODO: Set retry from config
				return "", backoff.RetryAfter(10)
			}

			if token != nil {
				if messaging.IsUnregistered(err) {
					// TODO: On second thought, this should be on some kind of queing system, so that
					// it will not disrupt an ongoing notification push
					err := notif.notifRepo.DeleteFCMToken(ctx, token.MemberID, token.MemberRegistrationBranchID, token.DeviceID)
					if err != nil {
						// TODO: Proper logging
						notif.logger.Error(fmt.Sprintf("failed to delete invalid FCM token: %s", err.Error()))
					}
				}

				err := notif.notifRepo.DeleteMemberNotification(ctx, token.MemberID, token.MemberRegistrationBranchID, msgModel.ID)
				if err != nil {
					return "", backoff.Permanent(err)
				}
			}

			return "", backoff.Permanent(err)
		}

		if token != nil {
			memberNotifModel.FCMMessageID = &fcmId
			memberNotifModel.Status = notifSent
			err = notif.notifRepo.SaveMemberNotification(ctx, memberNotifModel)
			if err != nil {
				return "", backoff.Permanent(err)
			}
		}

		return fcmId, nil
	}),
		backoff.WithBackOff(notif.retryBackoff),
		// TODO: Set this value from config
		backoff.WithMaxTries(3),
		backoff.WithNotify(func(err error, _ time.Duration) {
			logMsg, logAttrs := logAttrsFailToSendMsg(msgModel, token, err)
			notif.logger.Error(logMsg, logAttrs...)
		}),
	)

	if err != nil {
		// TODO: Check if error is invalid FCM token, and if it's invalid token,
		// log with WARN level
		if uErr := errors.Unwrap(err); uErr != nil {
			err = uErr
		}

		logMsg, logAttrs := logAttrsFailToSendMsg(msgModel, token, err)
		notif.logger.Error(logMsg, logAttrs...)
	}
}

func (notif *Notification) submitNotifPushWorker(fcmMsg *messaging.Message, notifMsgModel NotificationMessageModel, token *FCMToken) error {
	for {
		err := notif.gPool.Submit(func() {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Duration(notif.cfg.SendTimeout))
			notif.push(ctx, fcmMsg, notifMsgModel, token)
			defer cancel()
		})

		if err != nil {
			if err == ants.ErrPoolOverload {
				// Continue to try to submit until we can snatch
				// a worker
				continue
			}

			return err
		}

		return nil
	}
}

func (notif *Notification) _pushMsg(msg Message) {
	ctx := context.Background()
	msgData := msg.Action.Data
	notifMsgModel := NotificationMessageModel{
		Topic:    msg.Action.Topic,
		Title:    msg.Title,
		Body:     msg.Body,
		ImageURL: msg.ImageURL,
		Data: NotificationDataModel{
			Screen: msgData.Screen,
			URL:    msgData.URL,
		},
	}

	notifMsgModel, err := notif.notifRepo.CreateNotificationMessage(ctx, notifMsgModel)
	if err != nil {
		notif.logger.Error(err.Error())
		return
	}

	if msg.Action.TargetFilter == nil {
		fcmMsg := notif.buildFcmMessage(notifMsgModel.ID, msg, nil)
		err := notif.submitNotifPushWorker(fcmMsg, notifMsgModel, nil)
		if err != nil {
			// TODO: Log the error
		}

		return
	}

	pagination := NewSimplePagination(notif.cfg.ConcurrentPushLimit)
	for {
		tokens, err := notif.notifRepo.GetFCMTokensByFilter(ctx, *msg.Action.TargetFilter, pagination)
		if err != nil {
			// TODO: Log the error
			return
		}

		if len(tokens) == 0 {
			return
		}

		for _, token := range tokens {
			fcmMsg := notif.buildFcmMessage(notifMsgModel.ID, msg, &token)
			err := notif.submitNotifPushWorker(fcmMsg, notifMsgModel, &token)
			if err != nil {
				// TODO: Log the error

				return
			}
		}

		pagination = pagination.Next()
	}
}

func (notif *Notification) buildFcmMessage(notifId int, msg Message, token *FCMToken) *messaging.Message {
	fcmMsgData := make(fcmNotificationData)
	fcmMsgData.setNotificationId(notifId)
	fcmMsgData.setScreen(topicToMobileUiRouting[msg.Action.Topic])

	fcmMsg := messaging.Message{
		Notification: &messaging.Notification{
			Title:    msg.Title,
			Body:     msg.Body,
			ImageURL: msg.ImageURL,
		},
		Data: fcmMsgData,
	}

	if msg.Action.Data != nil {
		fcmMsgData.setUrl(msg.Action.Data.URL)
	}

	if token != nil {
		fcmMsg.Token = token.Token
	} else {
		fcmMsg.Topic = topicToFcmMsgTopic[msg.Action.Topic]
	}

	return &fcmMsg
}

func (notif *Notification) Close() error {
	notif.isClosed.Store(true)
	err := notif.gPool.ReleaseTimeout(time.Duration(notif.cfg.WaitWorkerPoolFinishOnClose))
	return err
}

func isValidTopic(topic string) bool {
	switch topic {
	case topicGeneral, topicPromo:
		return true
	}

	return false
}
