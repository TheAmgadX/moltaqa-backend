package auth

import (
	"errors"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAccessTokenTTL = 15 * time.Minute
)

// JWTSigner signs and verifies stateless access tokens.
type JWTSigner struct {
	signingKey []byte
	issuer     string
	ttl        time.Duration
}

func NewJWTSigner() *JWTSigner {
	return &JWTSigner{
		signingKey: []byte(env.GetString("JWT_SIGNING_KEY", "auth-secret")),
		issuer:     env.GetString("JWT_ISSUER", "auth-service"),
		ttl:        defaultAccessTokenTTL,
	}
}

func (s *JWTSigner) Sign(subject string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.signingKey)
}

func (s *JWTSigner) Verify(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return s.signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
