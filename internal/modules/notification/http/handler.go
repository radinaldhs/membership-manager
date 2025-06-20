package http

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	jwtfromctx "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/jwt_from_ctx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/http"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/notification"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type Handler struct {
	logger        *logging.Logger
	notifProvider notification.NotificationProvider
	cfg           config.Config
}

func NewHandler(cfg config.Config, logger *logging.Logger, notifProvider notification.NotificationProvider) *Handler {
	return &Handler{
		logger:        logger,
		notifProvider: notifProvider,
		cfg:           cfg,
	}
}

func (h *Handler) RegisterRoute(router *echo.Echo) {
	featureAccessAuthMddl := http.NewAuthMiddleware(h.cfg.Auth.JWT, h.logger, services.JWTScopeFeatures)
	adminAuthMddl := http.NewAdminAuthMiddleware(h.cfg.Auth, h.logger)

	group := router.Group("/notifications", featureAccessAuthMddl)
	group.POST("/member_fcm_tokens", h.RegisterFCMToken)
	group.POST("/status", h.SetMemberNotificationStatus)
	group.GET("", h.GetMemberNotificationList)
	group.DELETE("/:notif_id", h.DeleteMemberNotification)

	// Actions that can only be performed by admins
	admin := router.Group("/admin/notifications", adminAuthMddl)
	admin.POST("/push", h.PushNotification)
}

func (h *Handler) RegisterFCMToken(ectx echo.Context) error {
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	ctx := ectx.Request().Context()
	var req notification.FCMToken
	if err := ectx.Bind(&req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	req.MemberID = idPair.MemberID
	req.MemberRegistrationBranchID = idPair.RegistrationBranchID

	// TODO: Validate request

	if err := h.notifProvider.SaveFCMToken(ctx, req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Success", nil)
}

func (h *Handler) PushNotification(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	var req notification.Message
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	// TODO: Validate request

	if err := h.notifProvider.Push(ctx, req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Success", nil)
}

func (h *Handler) SetMemberNotificationStatus(ectx echo.Context) error {
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	ctx := ectx.Request().Context()
	var req notification.ClientNotificationStatus
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	err = h.notifProvider.SetMemberNotificationStatus(ctx, idPair.MemberID, idPair.RegistrationBranchID, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Success", nil)
}

func (h *Handler) GetMemberNotificationList(ectx echo.Context) error {
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	ctx := ectx.Request().Context()
	pageStr := ectx.QueryParam("page")
	pageSizeStr := ectx.QueryParam("page_size")
	if pageStr == "" {
		pageStr = "1"
	}

	if pageSizeStr == "" {
		pageSizeStr = "0"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, errors.New(customerror.ErrGroupClientErr, "", "invalid value for query page: not an integer"))
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, errors.New(customerror.ErrGroupClientErr, "", "invalid value for query page_size: not an integer"))
	}

	notifList, err := h.notifProvider.GetMemberNotificationList(ctx, idPair.MemberID, idPair.RegistrationBranchID, page, pageSize)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", notifList)
}

func (h *Handler) DeleteMemberNotification(ectx echo.Context) error {
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	ctx := ectx.Request().Context()

	notifIdStr := ectx.Param("notif_id")
	notifId, err := strconv.Atoi(notifIdStr)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, errors.New(customerror.ErrGroupClientErr, "", "invalid value for notif_id: not an integer"))
	}

	err = h.notifProvider.DeleteMemberNotification(ctx, idPair.MemberID, idPair.RegistrationBranchID, notifId)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Success", nil)
}
