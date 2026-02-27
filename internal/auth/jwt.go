package auth

import (
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator interface {
	Parse(token string) (*jwt.Token, error)
}

type jwksValidator struct{ jwks *keyfunc.JWKS }

func NewJWTValidator(jwksURL string) (JWTValidator, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshTimeout:    5 * time.Second,
		RefreshRateLimit:  time.Minute,
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}
	return &jwksValidator{jwks: jwks}, nil
}

func (v *jwksValidator) Parse(tok string) (*jwt.Token, error) {
	return jwt.Parse(tok, v.jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256", "ES256"}))
}

func ExtractClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(token, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func ExtractClaimString(token, name string) (string, error) {
	claims, err := ExtractClaims(token)
	if err != nil {
		return "", err
	}
	if v, ok := claims[name]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	return "", errors.New("claim not found")
}
