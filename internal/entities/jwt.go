package entities

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}
