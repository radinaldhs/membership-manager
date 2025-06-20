package http

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"regexp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/entities"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/encryption"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/httpresp"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

var (
	bearerTokenRegex     = regexp.MustCompile(`^Bearer (?P<token>\S+)$`)
	tokenRegexGroupIndex = bearerTokenRegex.SubexpIndex("token")
)

var (
	errInvalidToken    = errors.New(customerror.ErrGroupUnauthorized, "", "Invalid token")
	errForbiddenAccess = errors.New(customerror.ErrGroupForbidden, "", "Access forbidded for this resource")
)

func getTokenFromBearer(ectx echo.Context) string {
	matches := bearerTokenRegex.FindStringSubmatch(ectx.Request().Header.Get("Authorization"))
	if len(matches) == 0 {
		return ""
	}

	token := matches[tokenRegexGroupIndex]

	return token
}

func NewAuthMiddleware(jwtCfg config.JWT, logger *logging.Logger, requireScope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenStr := getTokenFromBearer(c)
			var claims entities.JWTClaims
			_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
				return jwtCfg.SigningKey.Decoded, nil
			},
				jwt.WithAudience(jwtCfg.Audience),
				jwt.WithIssuer(jwtCfg.Issuer),
			)

			if err != nil {
				switch {
				case errors.Is(err, jwt.ErrTokenMalformed), errors.Is(err, jwt.ErrTokenUnverifiable),
					errors.Is(err, jwt.ErrTokenSignatureInvalid), errors.Is(err, jwt.ErrTokenExpired),
					errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
					return httpresp.HandleError(c, logger, errInvalidToken)
				}

				return httpresp.HandleError(c, logger, err)
			}

			if requireScope != "" {
				if claims.Scope != requireScope {
					return httpresp.HandleError(c, logger, errForbiddenAccess)
				}
			}

			c.Set("user", claims.Subject)

			return next(c)
		}
	}
}

func NewAdminAuthMiddleware(cfg config.Auth, logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract API key from Authorization header,
			// it should be sent as bearer token
			matches := bearerTokenRegex.FindStringSubmatch(c.Request().Header.Get("Authorization"))
			if len(matches) == 0 {
				return c.JSON(http.StatusUnauthorized, httpresp.Response{
					Message: "Invalid or missing API key",
				})
			}

			tokenBase64 := matches[tokenRegexGroupIndex]
			token, err := base64.RawStdEncoding.DecodeString(tokenBase64)
			if err != nil {
				internalErr := customerror.ErrInternal.WithError(err).WithSource()
				return httpresp.HandleError(c, logger, internalErr)
			}

			hashedToken, err := encryption.HMAC256Hash(cfg.AdminAPIKeyHashSecret.Decoded, token)
			if err != nil {
				internalErr := customerror.ErrInternal.WithError(err).WithSource()
				return httpresp.HandleError(c, logger, internalErr)
			}

			if !bytes.Equal(hashedToken, cfg.AdminAPIKeyHash.Decoded) {
				return c.JSON(http.StatusUnauthorized, httpresp.Response{
					Message: "Invalid API key",
				})
			}

			return next(c)
		}
	}
}
