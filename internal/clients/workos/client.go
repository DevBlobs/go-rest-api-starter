package workos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "https://api.workos.com"
const defaultProvider = "authkit"

type Client interface {
	AuthURL(redirectURI, state string, opts AuthOpts) (string, error)
	ExchangeCode(ctx context.Context, code, redirectURI string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	JWKSURL() string
	LogoutURL(sessionID, returnTo string) string
}

type AuthOpts struct {
	OrganizationID string
	ConnectionID   string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type User struct {
	Object            string                 `json:"object"`
	ID                string                 `json:"id"`
	Email             string                 `json:"email"`
	FirstName         *string                `json:"first_name"`
	LastName          *string                `json:"last_name"`
	EmailVerified     *bool                  `json:"email_verified"`
	ProfilePictureURL *string                `json:"profile_picture_url"`
	LastSignInAt      *string                `json:"last_sign_in_at"`
	ExternalID        *string                `json:"external_id"`
	Metadata          map[string]interface{} `json:"metadata"`
	Locale            *string                `json:"locale"`
	CreatedAt         *string                `json:"created_at"`
	UpdatedAt         *string                `json:"updated_at"`
}

type clientImpl struct {
	clientID     string
	clientSecret string
}

func New(cfg *Config) Client {
	return &clientImpl{
		clientID:     cfg.ClientID,
		clientSecret: cfg.APIKey,
	}
}

func (c *clientImpl) AuthURL(redirectURI, state string, opts AuthOpts) (string, error) {
	u, err := url.Parse(baseURL + "/user_management/authorize")
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)

	switch {
	case opts.ConnectionID != "":
		q.Set("connection_id", opts.ConnectionID)
	case opts.OrganizationID != "":
		q.Set("organization_id", opts.OrganizationID)
	default:
		q.Set("provider", defaultProvider)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *clientImpl) ExchangeCode(ctx context.Context, code, redirectURI string) (*Tokens, error) {
	form := url.Values{
		"grant_type":    []string{"authorization_code"},
		"code":          []string{code},
		"client_id":     []string{c.clientID},
		"client_secret": []string{c.clientSecret},
		"redirect_uri":  []string{redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/user_management/authenticate", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", resp.Status)
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Tokens{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    out.ExpiresIn,
	}, nil
}

func (c *clientImpl) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	form := url.Values{
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{refreshToken},
		"client_id":     []string{c.clientID},
		"client_secret": []string{c.clientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/user_management/authenticate", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed: %s", resp.Status)
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Tokens{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    out.ExpiresIn,
	}, nil
}

func (c *clientImpl) JWKSURL() string {
	return fmt.Sprintf("%s/sso/jwks/%s", baseURL, c.clientID)
}

func (c *clientImpl) LogoutURL(sessionID, returnTo string) string {
	u, _ := url.Parse(baseURL + "/user_management/sessions/logout")
	q := u.Query()
	q.Set("session_id", sessionID)
	q.Set("return_to", returnTo)
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *clientImpl) GetUser(ctx context.Context, userID string) (*User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("userID is required")
	}
	urlStr := fmt.Sprintf("%s/user_management/users/%s", baseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.clientSecret))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user failed: %s", resp.Status)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &u, nil
}
