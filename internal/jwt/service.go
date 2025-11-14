package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Service struct {
	config Config
}

func New(config Config) *Service {
	if config.AccessTokenTTL == 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL == 0 {
		config.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "default-issuer"
	}

	return &Service{
		config: config,
	}
}

func (s *Service) GenerateAccessToken(userID, name string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", NewError("TOKEN_GENERATION_FAILED", "Failed to generate access token", err)
	}

	return tokenString, nil
}

func (s *Service) GenerateRefreshToken(userID, name string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", NewError("TOKEN_GENERATION_FAILED", "Failed to generate refresh token", err)
	}

	return tokenString, nil
}

func (s *Service) GenerateTokenPair(userID, name string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, name)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(userID, name)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, ErrExpiredToken
		}
		return nil, NewError("TOKEN_VALIDATION_FAILED", "Failed to validate token", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *Service) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(claims.UserID, claims.Name)
}
