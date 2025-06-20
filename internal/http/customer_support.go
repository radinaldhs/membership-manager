package http

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type CustomerSupportHandler struct {
	logger             *logging.Logger
	customerSupportSvc services.CustomerSupportService
	validate           *validator.Validate
	translator         ut.Translator
}

func NewCustomerSupportHandler(logger *logging.Logger, validate *validator.Validate, translator ut.Translator, customerSupportSvc services.CustomerSupportService) *CustomerSupportHandler {
	return &CustomerSupportHandler{
		logger:             logger,
		validate:           validate,
		translator:         translator,
		customerSupportSvc: customerSupportSvc,
	}
}

func (h *CustomerSupportHandler) GetAdminWhatsApp(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	res, err := h.customerSupportSvc.GetAdminWhatsApp(ctx)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", res)
}
