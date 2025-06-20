package http

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type RouterConfig struct {
	MainRouter             *echo.Echo
	AuthHandler            *AuthHandler
	MemberHandler          *MemberHandler
	CustomerSupportHandler *CustomerSupportHandler
	Logger                 *logging.Logger
}

func BuildRoutes(appCfg config.Config, routerCfg RouterConfig) error {
	router := routerCfg.MainRouter

	featureAccessAuthMddl := NewAuthMiddleware(appCfg.Auth.JWT, routerCfg.Logger, services.JWTScopeFeatures)
	emailLoginAuthMddl := NewAuthMiddleware(appCfg.Auth.JWT, routerCfg.Logger, services.JWTScopeEmailSignin)

	authGroup := router.Group("/auth")
	authGroup.POST("/login", routerCfg.AuthHandler.Login)
	authGroup.GET("/time_until_login_unlock/:member_code", routerCfg.AuthHandler.GetTimeUntilLoginUnlock)
	authGroup.POST("/token/refresh", routerCfg.AuthHandler.AuthTokenFromRefreshToken)
	authGroup.POST("/app_account", routerCfg.AuthHandler.RegisterEmailAndPassword, featureAccessAuthMddl)
	authGroup.POST("/app_account/validate_verify_otp", routerCfg.AuthHandler.ValidateOTPForEmailAndPasswordVerification, featureAccessAuthMddl)
	authGroup.POST("/login_with_email", routerCfg.AuthHandler.LoginWithEmailAndPassword, emailLoginAuthMddl)
	authGroup.POST("/app_account/google_id", routerCfg.AuthHandler.RegisterEmailFromGoogleSignin, featureAccessAuthMddl)
	authGroup.POST("/app_account/google_signin", routerCfg.AuthHandler.LoginWithGoogleSignin, emailLoginAuthMddl)

	memberGroup := router.Group("/members")
	memberGroup.POST("/verify", routerCfg.AuthHandler.VerifyMember)
	memberGroup.GET("/me", routerCfg.MemberHandler.GetMember, featureAccessAuthMddl)
	memberGroup.PATCH("/me", routerCfg.MemberHandler.UpdateMember, featureAccessAuthMddl)

	router.GET("/whatsapp/support_message", routerCfg.CustomerSupportHandler.GetAdminWhatsApp)

	return nil
}
