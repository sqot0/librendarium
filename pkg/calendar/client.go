package calendar

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"librendarium/pkg/config"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenURL      = "https://oauth2.googleapis.com/token"
	calendarScope = "https://www.googleapis.com/auth/calendar"
)

// Client wraps a Google Calendar service-account token for event creation.
type Client struct {
	httpClient  *http.Client
	accessToken string
	calendarID  string
}

// NewClient builds a calendar client and exchanges the service-account JWT for an access token.
func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	token, err := fetchAccessToken(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		httpClient:  http.DefaultClient,
		accessToken: token,
		calendarID:  cfg.GoogleCalendarID,
	}, nil
}

func fetchAccessToken(ctx context.Context, cfg config.Config) (string, error) {
	privateKey, err := parsePrivateKey(cfg.PrivateKey())
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"iss":   cfg.GoogleCalendarClientEmail,
		"scope": calendarScope,
		"aud":   tokenURL,
		"exp":   now + 3600,
		"iat":   now,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	values := url.Values{}
	values.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	values.Set("assertion", signed)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("token endpoint %s: %s", resp.Status, string(body))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if payload.AccessToken == "" {
		return "", fmt.Errorf("token response missing access_token")
	}

	return payload.AccessToken, nil
}

func parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM data")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return rsaKey, nil
}
