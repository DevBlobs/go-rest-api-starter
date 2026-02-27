package testsuite

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/boilerplate-api/go-rest-api-starter/internal/clients/workos"

	"github.com/golang-jwt/jwt/v5"
)

// ---- WorkOS mock ----

type workOSMock struct {
	jwksURL string
	privKey *rsa.PrivateKey
}

var testAccessToken string
var testRefreshToken = "test-refresh"

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

var currentWorkOSMock *workOSMock

func newWorkOSMock() workos.Client {
	// Create a minimal RSA keypair and expose its public key via a JWKS endpoint.
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := key.PublicKey

	// Base64 URL encode without padding as required by JWK spec.
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	jwks := jwksResponse{Keys: []jwk{{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   n,
		E:   e,
	}}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))

	return &workOSMock{jwksURL: srv.URL, privKey: key}
}

func (w *workOSMock) signToken(sub string) string {
	// Convert simple test IDs to valid UUID format
	var uuidSub string
	switch sub {
	case "user_123":
		uuidSub = "00000000-0000-0000-0000-000000000123"
	default:
		uuidSub = sub
	}
	claims := jwt.MapClaims{
		"sub":   uuidSub,
		"email": "test@example.com",
		"iss":   "https://api.workos.com/user_management/client_test_client_id",
		"aud":   "test_client_id",
		"sid":   "session_123",
		"exp":   time.Now().Add(60 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = "test-key"
	tok, _ := t.SignedString(w.privKey)
	return tok
}

func (w *workOSMock) AuthURL(redirectURI, state string, opts workos.AuthOpts) (string, error) {
	return redirectURI + "?state=" + state, nil
}

func (w *workOSMock) ExchangeCode(ctx context.Context, code, redirectURI string) (*workos.Tokens, error) {
	if testAccessToken == "" {
		testAccessToken = w.signToken("user_123")
	}
	return &workos.Tokens{AccessToken: testAccessToken, RefreshToken: testRefreshToken, ExpiresIn: 3600}, nil
}

func (w *workOSMock) Refresh(ctx context.Context, refreshToken string) (*workos.Tokens, error) {
	// Issue a new token on refresh
	testAccessToken = w.signToken("user_123")
	return &workos.Tokens{AccessToken: testAccessToken, RefreshToken: testRefreshToken, ExpiresIn: 3600}, nil
}

func (w *workOSMock) JWKSURL() string { return w.jwksURL }

func (w *workOSMock) LogoutURL(sessionID, returnTo string) string { return returnTo }

func (w *workOSMock) GetUser(ctx context.Context, userID string) (*workos.User, error) {
	// Return a deterministic user based on the provided ID
	first := "Test"
	last := "User"
	return &workos.User{ID: userID, Email: "test@example.com", FirstName: &first, LastName: &last}, nil
}

// GetAuthCookies returns cookies for authenticated requests in tests.
func GetAuthCookies() []*http.Cookie {
	if currentWorkOSMock == nil {
		panic("workOSMock not initialized - call buildExternalTestDeps first")
	}
	// Always generate a fresh token with current issuer format
	testAccessToken = currentWorkOSMock.signToken("user_123")
	return []*http.Cookie{
		{Name: "access_token", Value: testAccessToken},
		{Name: "refresh_token", Value: testRefreshToken},
	}
}

// buildExternalTestDeps constructs mocks for all external clients used by the app.
func buildExternalTestDeps() workos.Client {
	wk := newWorkOSMock().(*workOSMock)
	currentWorkOSMock = wk
	return wk
}
