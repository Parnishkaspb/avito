package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Config struct {
	SecretKey       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}
