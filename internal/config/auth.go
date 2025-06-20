package config

import "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/config"

type JWT struct {
	SigningKey             config.RawBase64Encoded `env:"JWT_SIGNING_KEY"`
	Issuer                 string                  `env:"JWT_ISSUER" default:"https://tokomasjawa.com"`
	Expiration             config.HourDuration     `env:"JWT_EXPIRATION" default:"24"`
	RefreshTokenSigningKey config.RawBase64Encoded `env:"REFRESH_JWT_SIGNING_KEY"`
	RefreshTokenExpiration config.HourDuration     `env:"REFRESH_JWT_EXPIRATION" default:"48"`
	Audience               string                  `env:"JWT_AUDIENCE" default:"https://tokomasjawa.com"`
}

type OTP struct {
	Length                       int                   `env:"OTP_LENGTH"`
	ExpireDuration               config.MinuteDuration `env:"OTP_EXPIRE_DURATION"`
	RegenerationCooldownDuration config.MinuteDuration `env:"OTP_REGENERATION_COOLDOWN_DURATION" default:"5"`
}

type Google struct {
	ClientID string `env:"GOOGLE_CLIENT_ID"`
}

type Auth struct {
	CardNumberEncryptionKey             config.RawBase64Encoded `env:"CARD_NUMBER_ENCRYPTION_KEY"`
	WaitTimeMinutesOnFailedLogin5Times  config.MinuteDuration   `env:"WAIT_TIME_MINUTE_ON_FAILED_LOGIN_5_TIMES" default:"5"`
	WaitTimeMinutesOnFailedLogin15Times config.MinuteDuration   `env:"WAIT_TIME_MINUTE_ON_FAILED_LOGIN_15_TIMES" default:"30"`
	AdminAPIKeyHashSecret               config.RawBase64Encoded `env:"ADMIN_API_KEY_HASH_SECRET"`
	AdminAPIKeyHash                     config.RawBase64Encoded `env:"ADMIN_API_KEY_HASH"`
	JWT                                 JWT
	OTP                                 OTP
	Google                              Google
}
