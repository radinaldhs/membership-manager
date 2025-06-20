package models

import "time"

type VerifyMemberRequest struct {
	MemberCode string `json:"KodeMember"`
}

type VerifyMemberResponse struct {
	MemberCode  string `json:"KodeMember"`
	Name        string `json:"Nama"`
	PhoneNumber string `json:"Telpon"`
}

type LoginRequest struct {
	MemberCode  string `json:"KodeMember" validate:"required"`
	PhoneNumber string `json:"phone" validate:"required,phone_num"`
}

type AuthToken struct {
	Token                 string    `json:"token"`
	TokenExpiredAt        time.Time `json:"token_expired_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiredAt time.Time `json:"refresh_token_expired_at"`
}

type AuthTokenForLoginWithEmail struct {
	Token          string    `json:"token"`
	TokenExpiredAt time.Time `json:"token_expired_at"`
}

type LoginResponse struct {
	EmailVerified bool `json:"email_verified"`
	AuthToken
}

type GetTimeUntilLoginUnlockResponse struct {
	Until string `json:"until"`
}

type TokenFromRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RegisterEmailAndPasswordRequest struct {
	Email string `json:"email" validate:"required"`
}

type OTPDetail struct {
	Digits           int       `json:"otp_digits"`
	ExpiredAt        time.Time `json:"otp_expired_at"`
	NextRegeneration time.Time `json:"otp_next_regeneration"`
}

type OTP struct {
	OTP string `json:"otp"`
}

type ValidateOTPForEmailAndPasswordRegistrationRequest struct {
	OTP      string `json:"otp" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginWithEmailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterEmailFromGoogleSigninRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type LoginWithGoogleSigninRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}
