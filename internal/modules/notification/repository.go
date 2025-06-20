package notification

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type SimplePagination struct {
	size        int
	fetchOffset int
}

func NewSimplePagination(size int) SimplePagination {
	return SimplePagination{size: size}
}

func (sp SimplePagination) Next() SimplePagination {
	return SimplePagination{
		size:        sp.size,
		fetchOffset: sp.fetchOffset + sp.size,
	}
}

func (sp SimplePagination) Size() int {
	return sp.size
}

func (sp SimplePagination) Offset() int {
	return sp.fetchOffset
}

// TODO: Add validation
// TODO: It is probably best to create separate struct
// to handle insert and fetch from database
type FCMToken struct {
	MemberID                   int    `db:"member_id" json:"member_id"`
	MemberRegistrationBranchID int    `db:"member_regist_branch_id" json:"member_regist_branch_id"`
	DeviceID                   string `db:"device_id" json:"device_id"`
	Platform                   string `db:"platform" json:"platform"`
	Token                      string `db:"token" json:"token"`
}

type NotificationDataModel struct {
	Screen string `json:"screen"`
	URL    string `json:"url"`
}

func (notifData *NotificationDataModel) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, notifData)

	case string:
		return json.Unmarshal([]byte(v), notifData)
	}

	return errors.New("supported data type to scan: []byte, string")
}

func (notifData NotificationDataModel) Value() (driver.Value, error) {
	return json.Marshal(notifData)
}

type NotificationMessageModel struct {
	ID       int                   `db:"id"`
	Topic    string                `db:"topic"`
	Title    string                `db:"title"`
	Body     string                `db:"body"`
	ImageURL string                `db:"image_url"`
	Data     NotificationDataModel `db:"data"`
}

type MemberNotificationModel struct {
	MemberID                   int     `db:"member_id"`
	MemberRegistrationBranchID int     `db:"member_regist_branch_id"`
	NotificationID             int     `db:"notification_id"`
	FCMMessageID               *string `db:"fcm_message_id"`
	Status                     string  `db:"status"`
}

type MemberNotificationListItemViewModel struct {
	NotificationID int     `db:"notification_id" json:"notification_id"`
	FCMMessageID   *string `db:"fcm_message_id" json:"fcm_message_id"`
	Status         string  `db:"status" json:"status"`
	Title          string  `db:"title" json:"title"`
	Body           string  `db:"body" json:"body"`
	ImageURL       string  `db:"image_url" json:"image_url"`
}

type NotificationRepository interface {
	SaveFCMToken(ctx context.Context, token FCMToken) error
	GetFCMTokensByFilter(ctx context.Context, filter TargetFilter, pagination SimplePagination) ([]FCMToken, error)
	CreateNotificationMessage(ctx context.Context, msg NotificationMessageModel) (NotificationMessageModel, error)
	SaveMemberNotification(ctx context.Context, memberNotif MemberNotificationModel) error
	DeleteMemberNotification(ctx context.Context, memberId, memberRegistBranchId, notifId int) error
	GetMemberNotifications(ctx context.Context, memberId, memberRegistBranchId int, offset, limit int) ([]MemberNotificationListItemViewModel, error)
	DeleteFCMToken(ctx context.Context, memberId, memberRegistBranchId int, deviceId string) error
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) SaveFCMToken(ctx context.Context, token FCMToken) error {
	// NOTE: We are not using stored procedure anymore, and will probably not use it
	// in the future
	// q := `CALL member_UpdateMTokenFCM(?, ?, ?)`
	// _, err := repo.db.ExecContext(ctx, q, token.MemberID, token.MemberRegistrationBranchID, token.Token)
	// if err != nil {
	// 	return customerror.ErrDatabase.WithError(err).WithSource()
	// }

	q := `INSERT INTO member_mtokenfcm (
	IdMMember,
	IdMCabangDaftar,
	DeviceId,
	Platform,
	TokenFCM,
	TimeCreate,
	TimeUpdate,
	TimeLastUsed
) VALUES (
	:member_id,
	:member_regist_branch_id,
	:device_id,
	:platform,
	:token,
	NOW(),
	NOW(),
	NOW()
) ON duplicate KEY UPDATE
	TokenFCM = :token,
	Platform = :platform,
	TimeUpdate = NOW()`

	_, err := repo.db.NamedExecContext(ctx, q, token)

	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *Repository) GetFCMTokensByFilter(ctx context.Context, filter TargetFilter, pagination SimplePagination) ([]FCMToken, error) {
	var tokens []FCMToken
	q, args := getFcmTokensByFilterQuery(filter, pagination)
	rows, err := repo.db.QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, customerror.ErrDatabase.WithError(err).WithSource()
	}

	for rows.Next() {
		var token FCMToken
		if err := rows.StructScan(&token); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, customerror.ErrDatabase.WithError(err).WithSource()
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (repo *Repository) DeleteFCMToken(ctx context.Context, memberId, memberRegistBranchId int, deviceId string) error {
	q := `DELETE FROM member_mtokenfcm WHERE 
		IdMMember = ? 
		AND IdMCabangDaftar = ?
		AND DeviceId = ?`

	_, err := repo.db.ExecContext(ctx, q, memberId, memberRegistBranchId, deviceId)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *Repository) CreateNotificationMessage(ctx context.Context, msg NotificationMessageModel) (NotificationMessageModel, error) {
	q := `INSERT INTO member_mnotifikasi (
	Topic,
	Judul,
	Konten,
	URLGambar,
	DataJson,
	TimeCreate
) VALUES (
	:topic,
	:title,
	:body,
	:image_url,
	:data,
	NOW()
)`
	res, err := repo.db.NamedExecContext(ctx, q, msg)
	if err != nil {
		return NotificationMessageModel{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return NotificationMessageModel{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	msg.ID = int(lastId)

	return msg, nil
}

func (repo *Repository) SaveMemberNotification(ctx context.Context, memberNotif MemberNotificationModel) error {
	q := `INSERT INTO member_mnotifikasimember (
	IdMMember,
	IdMCabangDaftar,
	IdMNotifikasi,
	IdFCMMessage,
	StatusNotifikasi,
	TimeCreate
) VALUES (
	:member_id,
	:member_regist_branch_id,
	:notification_id,
	:fcm_message_id,
	:status,
	NOW()
) 
ON duplicate KEY UPDATE 
	IdFCMMessage = :fcm_message_id,
	StatusNotifikasi = :status,
	TimeUpdate = NOW()`

	_, err := repo.db.NamedExecContext(ctx, q, memberNotif)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *Repository) DeleteMemberNotification(ctx context.Context, memberId, memberRegistBranchId, notifId int) error {
	q := `DELETE FROM member_mnotifikasimember WHERE 
	IdMMember = ?
	AND IdMCabangDaftar = ?
	AND IdMNotifikasi = ?`

	_, err := repo.db.ExecContext(ctx, q, memberId, memberRegistBranchId, notifId)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *Repository) GetMemberNotifications(ctx context.Context, memberId, memberRegistBranchId int, offset, limit int) ([]MemberNotificationListItemViewModel, error) {
	q := `SELECT 
	mm.IdMNotifikasi AS notification_id,
	mm.IdFCMMessage AS fcm_message_id,
	mm.StatusNotifikasi AS status,
	mmn.Judul AS title,
	mmn.Konten AS body,
	mmn.URLGambar AS image_url
FROM member_mnotifikasimember mm 
	JOIN member_mnotifikasi mmn ON mmn.IdMNotifikasi = mm.IdMNotifikasi
WHERE 
	mm.IdMMember = ?
	AND mm.IdMCabangDaftar = ?
ORDER BY mm.TimeCreate DESC
LIMIT ?, ?`

	rows, err := repo.db.QueryxContext(ctx, q, memberId, memberRegistBranchId, offset, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, customerror.ErrDatabase.WithError(err).WithSource()
	}

	var list []MemberNotificationListItemViewModel

	for rows.Next() {
		var item MemberNotificationListItemViewModel
		if err := rows.StructScan(&item); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, customerror.ErrDatabase.WithError(err).WithSource()
		}

		list = append(list, item)
	}

	return list, nil
}

// TODO: It is high time for us to use SQL builder...
func getFcmTokensByFilterQuery(filter TargetFilter, pagination SimplePagination) (string, []any) {
	q := `SELECT
	mmt.IdMMember AS member_id,
	mmt.IdMCabangDaftar AS member_regist_branch_id,
	mmt.DeviceId AS device_id,
	mmt.Platform AS platform,
	mmt.TokenFCM AS token
FROM member_mtokenfcm mmt
JOIN member_mmember mm ON 
	mm.IdMMember = mmt.IdMMember
	AND mm.IdMCabangDaftar = mmt.IdMCabangDaftar
JOIN member_mkartumember mk ON 
	mk.IdMKartuMember = mm.IdMKartuMember
	AND mk.IsActive = 1
WHERE (DATEDIFF(NOW(), mmt.TimeLastUsed)) < 270`

	conditions := []string{q}
	var args []any

	if len(filter.MemberType) != 0 {
		var placeholder []string
		var params string
		for _, t := range filter.MemberType {
			placeholder = append(placeholder, "?")
			args = append(args, t)
		}

		params = strings.Join(placeholder, ",")
		conditions = append(conditions, fmt.Sprintf("mk.IdMTipeKartuMember IN (%s)", params))
	}

	if len(filter.Gender) != 0 {
		genders := filter.Gender
		if len(genders) > 2 {
			genders = genders[0:2]
		}

		var placeholder []string
		var params string
		for _, t := range genders {
			placeholder = append(placeholder, "?")
			args = append(args, t)
		}

		params = strings.Join(placeholder, ",")

		conditions = append(conditions, fmt.Sprintf("mm.Kelamin IN (%s)", params))
	}

	if len(filter.Age) == 1 {
		conditions = append(conditions, "(TIMESTAMPDIFF(YEAR, mm.TglLahir, NOW())) = ?")
		args = append(args, filter.Age[0])
	}

	if len(filter.Age) > 1 {
		conditions = append(conditions, "(TIMESTAMPDIFF(YEAR, mm.TglLahir, NOW())) BETWEEN ? AND ?")
		args = append(args, filter.Age[0], filter.Age[1])
	}

	if len(filter.MemberCodes) != 0 {
		var placeholder []string
		var params string
		for _, t := range filter.MemberCodes {
			placeholder = append(placeholder, "?")
			args = append(args, t)
		}

		params = strings.Join(placeholder, ",")
		conditions = append(conditions, fmt.Sprintf("(mk.NomorKartu IN (%s) AND mk.IsActive = TRUE)", params))
	}

	q = strings.Join(conditions, " AND ")
	q += fmt.Sprintf("LIMIT %d, %d", pagination.Offset(), pagination.Size())

	return q, args
}
