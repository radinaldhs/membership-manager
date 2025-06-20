package http

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	jwtfromctx "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/jwt_from_ctx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/models"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type MemberHandler struct {
	logger     *logging.Logger
	memberSvc  services.MemberService
	validate   *validator.Validate
	translator ut.Translator
}

func NewMemberHandler(logger *logging.Logger, validate *validator.Validate, translator ut.Translator, memberSvc services.MemberService) *MemberHandler {
	return &MemberHandler{
		logger:     logger,
		memberSvc:  memberSvc,
		validate:   validate,
		translator: translator,
	}
}

func (h *MemberHandler) GetMember(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	member, err := h.memberSvc.GetMember(ctx, idPair.MemberID, idPair.RegistrationBranchID)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", member)
}

func (h *MemberHandler) UpdateMember(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	var req models.UpdateMemberRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	if err := h.validate.Struct(req); err != nil {
		return httpresp.HandleValidationError(ectx, h.logger, h.validate, h.translator, err)
	}

	err = h.memberSvc.UpdateMember(ctx, idPair.MemberID, idPair.RegistrationBranchID, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Berhasil merubah profil", nil)
}
