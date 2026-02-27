package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	stateTTL = 10 * time.Minute
)

type StateClaims struct {
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}

func GenerateState(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("state secret cannot be empty")
	}

	nonce := randomState()
	now := time.Now()

	claims := StateClaims{
		Nonce: nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(stateTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateState(secret, state string) error {
	if secret == "" {
		return errors.New("state secret cannot be empty")
	}

	token, err := jwt.ParseWithClaims(state, &StateClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return fmt.Errorf("invalid state token: %w", err)
	}

	if claims, ok := token.Claims.(*StateClaims); ok && token.Valid {
		if claims.Nonce == "" {
			return errors.New("state token missing nonce")
		}
		return nil
	}

	return errors.New("invalid state token claims")
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
