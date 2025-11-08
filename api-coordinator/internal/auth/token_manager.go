package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenManager struct {
	secret []byte
}

// NewJWTTokenManager creates a new JWT token manager.
func NewJWTTokenManager(secret string) TokenManager {
	return &jwtTokenManager{
		secret: []byte(secret),
	}
}

func (j *jwtTokenManager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}
