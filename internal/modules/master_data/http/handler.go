package http

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/http"
	masterdata "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/master_data"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type Handler struct {
	logger     *logging.Logger
	masterData masterdata.MasterDataProvider
	cfg        config.Config
}

func NewHandler(cfg config.Config, logger *logging.Logger, masterData masterdata.MasterDataProvider) *Handler {
	return &Handler{
		logger:     logger,
		masterData: masterData,
		cfg:        cfg,
	}
}

func (h *Handler) RegisterRoute(router *echo.Echo) {
	featureAccessAuthMddl := http.NewAuthMiddleware(h.cfg.Auth.JWT, h.logger, services.JWTScopeFeatures)

	group := router.Group("masterdata", featureAccessAuthMddl)
	group.GET("/provinces", h.GetProvinces)
	group.GET("/provinces/:province_region_code/regencies", h.GetRegencies)
}

func (h *Handler) GetProvinces(ectx echo.Context) error {
	provinces, err := h.masterData.GetProvinces(ectx.Request().Context())
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", provinces)
}

func (h *Handler) GetRegencies(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	provinceRegionCode := ectx.Param("province_region_code")
	regencies, err := h.masterData.GetRegenciesByProvince(ctx, provinceRegionCode)
	if err != nil {
		return httpresp.HandleError(ectx, h.logger, err)
	}

	return httpresp.OKResponse(ectx, "", regencies)
}
