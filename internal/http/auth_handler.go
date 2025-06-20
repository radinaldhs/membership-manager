package http

import (
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	jwtfromctx "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/jwt_from_ctx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/models"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type AuthHandler struct {
	logger     *logging.Logger
	authSvc    services.AuthService
	validate   *validator.Validate
	translator ut.Translator
}

func NewAuthHandler(logger *logging.Logger, validate *validator.Validate, translator ut.Translator, authSvc services.AuthService) *AuthHandler {
	return &AuthHandler{
		logger:     logger,
		validate:   validate,
		translator: translator,
		authSvc:    authSvc,
	}
}

func (h *AuthHandler) VerifyMember(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	var req models.VerifyMemberRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	memberModel, err := h.authSvc.VerifyMember(ctx, req.MemberCode)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return ectx.JSON(http.StatusOK, memberModel)
}

func (h *AuthHandler) Login(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	var req models.LoginRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	if err := h.validate.StructCtx(ctx, req); err != nil {
		return httpresp.HandleValidationError(ectx, h.logger, h.validate, h.translator, err)
	}

	res, err := h.authSvc.Login(ctx, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Berhasil login", res)
}

func (h *AuthHandler) GetTimeUntilLoginUnlock(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	memberCode := ectx.Param("member_code")

	res, err := h.authSvc.GetTimeUntilLoginUnlock(ctx, memberCode)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", res)
}

func (h *AuthHandler) AuthTokenFromRefreshToken(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	var req models.TokenFromRefreshTokenRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	if err := h.validate.StructCtx(ctx, req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	res, err := h.authSvc.TokenFromRefreshToken(ctx, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", res)
}

func (h *AuthHandler) RegisterEmailAndPassword(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	var req models.RegisterEmailAndPasswordRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	resp, err := h.authSvc.RegisterEmailByOTP(ctx, idPair.MemberID, idPair.RegistrationBranchID, req.Email)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", resp)
}

func (h *AuthHandler) ValidateOTPForEmailAndPasswordVerification(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	var req models.ValidateOTPForEmailAndPasswordRegistrationRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	if err := h.validate.StructCtx(ctx, req); err != nil {
		return httpresp.HandleValidationError(ectx, h.logger, h.validate, h.translator, err)
	}

	err = h.authSvc.ValidateOTPForEmailAndPasswordRegistration(ctx, idPair.MemberID, idPair.RegistrationBranchID, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Sukses mendaftarkan email dan password", nil)
}

func (h *AuthHandler) LoginWithEmailAndPassword(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	var req models.LoginWithEmailAndPasswordRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	resp, err := h.authSvc.LoginWithEmailAndPassword(ctx, idPair.MemberID, idPair.RegistrationBranchID, req)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Berhasil login", resp)
}

func (h *AuthHandler) RegisterEmailFromGoogleSignin(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	var req models.RegisterEmailFromGoogleSigninRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	if err := h.validate.Struct(req); err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	resp, err := h.authSvc.RegisterEmailFromGoogleSignin(ctx, idPair.MemberID, idPair.RegistrationBranchID, req.IDToken)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "Sukses mendaftarkan email", resp)
}

func (h *AuthHandler) LoginWithGoogleSignin(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	idPair, err := jwtfromctx.GetMemberIdentityPair(ectx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	var req models.LoginWithGoogleSigninRequest
	if err := ectx.Bind(&req); err != nil {
		return httpresp.BadRequestResponse(ectx)
	}

	resp, err := h.authSvc.LoginWithGoogleSignin(ctx, idPair.MemberID, idPair.RegistrationBranchID, req.IDToken)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", resp)
}
