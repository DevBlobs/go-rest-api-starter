package auth

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type meResponse struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     string  `json:"email"`
}

const (
	GrantTypeAuthorizationCode = "authorization_code"
)

type tokenRequest struct {
	GrantType   string `form:"grant_type" validate:"required"`
	Code        string `form:"code" validate:"required"`
	RedirectURI string `form:"redirect_uri"`
	State       string `form:"state" validate:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Handler struct {
	Svc         Service
	Cookie      CookieCfg
	BaseURL     string
	LoginOpts   Opts
	StateSecret string
}

func (ctl *Handler) RegisterPublic(g *echo.Group) {
	g.GET("/login", ctl.Login)
	g.GET("/callback", ctl.Callback)
	g.POST("/logout", ctl.Logout)
	g.POST("/token", ctl.Token)
}

func (ctl *Handler) RegisterProtected(g *echo.Group) {
	g.GET("/me", ctl.Me)
}

func (ctl *Handler) Login(c echo.Context) error {
	state, err := GenerateState(ctl.StateSecret)
	if err != nil {
		slog.Error("failed to generate state", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "authentication configuration error")
	}

	u, err := ctl.Svc.LoginURL(state, ctl.LoginOpts)
	if err != nil {
		slog.Error("login failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, u)
}

func (ctl *Handler) Callback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing code")
	}

	state := c.QueryParam("state")
	if state == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing state")
	}

	if err := ValidateState(ctl.StateSecret, state); err != nil {
		slog.Error("invalid state", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	t, err := ctl.Svc.Exchange(c.Request().Context(), code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	setCookie(c.Response(), "access_token", t.AccessToken, ctl.Cookie, time.Duration(t.ExpiresIn-60)*time.Second, true)
	setCookie(c.Response(), "refresh_token", t.RefreshToken, ctl.Cookie, 30*24*time.Hour, true)

	return c.Redirect(http.StatusFound, ctl.BaseURL)
}

func (ctl *Handler) Logout(c echo.Context) error {

	delCookie(c.Response(), "access_token", ctl.Cookie)
	delCookie(c.Response(), "refresh_token", ctl.Cookie)

	sessionID := c.QueryParam("session_id")
	if sessionID == "" {
		if at, err := c.Cookie("access_token"); err == nil && at.Value != "" {
			if sid := extractSID(at.Value); sid != "" {
				sessionID = sid
			}
		}
	}

	returnTo := c.QueryParam("return_to")
	if returnTo == "" {
		returnTo = strings.TrimRight(ctl.BaseURL, "/") + "/signed-out"
	}

	if sessionID != "" {
		u := ctl.Svc.LogoutURL(sessionID, returnTo)
		return c.Redirect(http.StatusFound, u)
	}

	return c.NoContent(http.StatusNoContent)
}

func extractSID(jwtToken string) string {
	sid, err := ExtractClaimString(jwtToken, "sid")
	if err != nil {
		return ""
	}
	return sid
}

func (ctl *Handler) Me(c echo.Context) error {
	ctx := c.Request().Context()
	principal, ok := GetPrincipal(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "unauthenticated"})
	}

	actor, err := ctl.Svc.GetActor(ctx, principal.Subject)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch user profile")
	}

	return c.JSON(http.StatusOK, meResponse{
		FirstName: actor.FirstName,
		LastName:  actor.LastName,
		Email:     actor.Email,
	})
}

func (ctl *Handler) Token(c echo.Context) error {
	var req tokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.GrantType != GrantTypeAuthorizationCode {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported grant type")
	}

	if err := ValidateState(ctl.StateSecret, req.State); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	t, err := ctl.Svc.Exchange(c.Request().Context(), req.Code)
	if err != nil {
		slog.Error("token exchange failed", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "token exchange failed")
	}

	return c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
		TokenType:    "Bearer",
	})
}
