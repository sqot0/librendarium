package librusapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"librendarium/pkg/config"
)

const (
	librusBaseURL  = "https://synergia.librus.pl/gateway/api/2.0/"
	oauthClientID  = "46"
	oauthBase      = "https://api.librus.pl/OAuth/Authorization"
	oauthGrant     = "https://api.librus.pl/OAuth/Authorization/Grant?client_id=" + oauthClientID
	oauthTokenInfo = "https://synergia.librus.pl/gateway/api/2.0/Auth/TokenInfo/"
	oauthScope     = "mydata"
)

// Client wraps the HTTP client that keeps Librus session cookies.
type Client struct {
	httpClient *http.Client
}

// NewClient authenticates against Libra and returns an authenticated HTTP client.
func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar: %w", err)
	}

	httpClient := &http.Client{Jar: jar}

	if err := performLogin(ctx, httpClient, cfg.LibrusLogin, cfg.LibrusPassword); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, oauthTokenInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare token info request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("librus token info failure (%s): %s", resp.Status, string(body))
	}

	return &Client{httpClient: httpClient}, nil
}

func performLogin(ctx context.Context, httpClient *http.Client, login, password string) error {
	testAuth := fmt.Sprintf("%s?client_id=%s&response_type=code&scope=%s", oauthBase, oauthClientID, oauthScope)
	if err := sendEmptyGet(ctx, httpClient, testAuth); err != nil {
		return err
	}

	form := url.Values{}
	form.Set("action", "login")
	form.Set("login", login)
	form.Set("pass", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthBase+"?client_id="+oauthClientID, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("submit login: %w", err)
	}

	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("read login response: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("librus login failed: %s", resp.Status)
	}

	return sendEmptyGet(ctx, httpClient, oauthGrant)
}

func sendEmptyGet(ctx context.Context, client *http.Client, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("GET request %s: %w", rawURL, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", rawURL, err)
	}
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("read login response: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request %s returned %s", rawURL, resp.Status)
	}
	return nil
}

func (c *Client) getAPI(ctx context.Context, endpoint string) ([]byte, error) {
	full := librusBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare %s request: %w", endpoint, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s fetch failed: %s %s", endpoint, resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s body: %w", endpoint, err)
	}
	return body, nil
}
