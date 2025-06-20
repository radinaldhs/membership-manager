package app

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/validator"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/http"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

func newHttpRouterConfig(logger *logging.Logger, serviceColl *serviceCollection) (http.RouterConfig, error) {
	router := echo.New()
	router.HideBanner = true
	validate, translator, err := validator.NewValidator()
	if err != nil {
		return http.RouterConfig{}, err
	}

	// TODO: Disable/enable debug mode from configuration

	return http.RouterConfig{
		MainRouter:             router,
		AuthHandler:            http.NewAuthHandler(logger, validate, translator, serviceColl.authSvc),
		MemberHandler:          http.NewMemberHandler(logger, validate, translator, serviceColl.memberSvc),
		CustomerSupportHandler: http.NewCustomerSupportHandler(logger, validate, translator, serviceColl.customerSupportSvc),
		Logger:                 logger,
	}, nil
}
