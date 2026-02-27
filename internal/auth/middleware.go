package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type tokenSource int

const (
	tokenSourceNone tokenSource = iota
	tokenSourceBearer
	tokenSourceCookie
)

type Middleware struct {
	Svc            Service
	Validator      JWTValidator
	Cookie         CookieCfg
	ClientID       string
	AllowedIssuers []string
	Namespace      string
}

func (m *Middleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, source := m.extractToken(c)
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
		}

		parsed, err := m.Validator.Parse(token)
		if err != nil || !parsed.Valid {
			// For cookie-based tokens, attempt refresh on expiry
			if source == tokenSourceCookie && isTokenExpiredError(err) {
				if refreshErr := m.tryRefreshForCookieBaseToken(c); refreshErr == nil {
					if ck, _ := c.Cookie("access_token"); ck.Value != "" {
						parsed, err = m.Validator.Parse(ck.Value)
					}
				}
			}
			if err != nil || !parsed.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
		}

		iss, hasIss := claims["iss"].(string)
		if !hasIss {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token issuer")
		}

		if !m.isAllowedIssuer(iss) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid issuer")
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing subject claim")
		}

		email := m.extractStringClaim(claims, "email")
		orgID := m.extractStringClaim(claims, "org_id")

		// Validate audience for M2M tokens only
		if orgID != "" {
			if !m.validateAudience(claims, m.ClientID) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid audience")
			}
		}

		scopes := m.parseScopes(claims)

		principalType := "user"
		if orgID != "" {
			principalType = "service"
		}

		principal := &Principal{
			Subject: sub,
			Type:    principalType,
			Email:   email,
			OrgID:   orgID,
			Scopes:  scopes,
			Claims:  claimsToMap(claims),
		}

		slog.Info("[AUTH]",
			"subject", principal.Subject,
			"type", principal.Type,
			"org_id", principal.OrgID,
			"scopes", formatScopes(principal.Scopes),
			"source", tokenSourceName(source),
		)

		c.Set(string(PrincipalContextKey), principal)

		return next(c)
	}
}

func (m *Middleware) extractToken(c echo.Context) (string, tokenSource) {
	if token, ok := extractBearerTokenFromAuthHeader(c); ok {
		return token, tokenSourceBearer
	}
	if token, ok := extractAccessTokenFromCookie(c); ok {
		return token, tokenSourceCookie
	}
	return "", tokenSourceNone
}

func (m *Middleware) tryRefreshForCookieBaseToken(c echo.Context) error {
	rt, err := c.Cookie("refresh_token")
	if err != nil || rt.Value == "" {
		return err
	}
	t, err := m.Svc.Refresh(c.Request().Context(), rt.Value)
	if err != nil {
		return err
	}
	setCookie(c.Response(), "access_token", t.AccessToken, m.Cookie, time.Duration(t.ExpiresIn-60)*time.Second, true)
	if t.RefreshToken != "" && t.RefreshToken != rt.Value {
		setCookie(c.Response(), "refresh_token", t.RefreshToken, m.Cookie, 30*24*time.Hour, true) // rotation
	}
	return nil
}

func (m *Middleware) validateAudience(claims jwt.MapClaims, expectedAud string) bool {
	switch aud := claims["aud"].(type) {
	case string:
		return aud == expectedAud
	case []string:
		for _, a := range aud {
			if a == expectedAud {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (m *Middleware) parseScopes(claims jwt.MapClaims) []string {
	scopeStr := m.extractStringClaim(claims, "scope")
	if scopeStr == "" {
		return []string{}
	}
	parts := strings.Fields(scopeStr)
	return parts
}

func (m *Middleware) extractStringClaim(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func RequireScope(requiredScope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			principal, ok := GetPrincipal(c)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
			}

			hasScope := false
			for _, s := range principal.Scopes {
				if s == requiredScope {
					hasScope = true
					break
				}
			}

			decision := "denied"
			if hasScope {
				decision = "granted"
			}

			slog.Info("[AUTHZ]",
				"endpoint", c.Request().URL.Path,
				"required_scope", requiredScope,
				"decision", decision,
			)

			if !hasScope {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient scope")
			}

			return next(c)
		}
	}
}

func isTokenExpiredError(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet)
}

func extractBearerTokenFromAuthHeader(c echo.Context) (string, bool) {
	authz := c.Request().Header.Get(echo.HeaderAuthorization)
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		token := strings.TrimSpace(authz[7:])
		if token != "" {
			return token, true
		}
	}
	return "", false
}

func extractAccessTokenFromCookie(c echo.Context) (string, bool) {
	ck, err := c.Cookie("access_token")
	if err == nil && ck != nil && ck.Value != "" {
		return ck.Value, true
	}
	return "", false
}

func claimsToMap(claims jwt.MapClaims) map[string]any {
	m := make(map[string]any, len(claims))
	for k, v := range claims {
		m[k] = v
	}
	return m
}

func (m *Middleware) isAllowedIssuer(iss string) bool {
	for _, allowed := range m.AllowedIssuers {
		if iss == allowed {
			return true
		}
	}
	return false
}

func formatScopes(scopes []string) string {
	if len(scopes) == 0 {
		return "[]"
	}
	return "[" + strings.Join(scopes, " ") + "]"
}

func tokenSourceName(ts tokenSource) string {
	switch ts {
	case tokenSourceBearer:
		return "bearer"
	case tokenSourceCookie:
		return "cookie"
	default:
		return "unknown"
	}
}
