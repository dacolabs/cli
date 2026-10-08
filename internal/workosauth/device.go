package workosauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const deviceCodeGrant = "urn:ietf:params:oauth:grant-type:device_code"

type Client struct {
	baseURL  string
	clientID string
	http     *http.Client
}

func New(baseURL, clientID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), clientID: clientID, http: httpClient}
}

type DeviceAuthorization struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               time.Duration
	Interval                time.Duration
}

type Tokens struct {
	AccessToken     string
	RefreshToken    string
	OrganizationID  string
	UserID          string
	UserEmail       string
	ExpiresAt       time.Time
}

type authError struct {
	Error string `json:"error"`
}

func (c *Client) AuthorizeDevice(ctx context.Context) (DeviceAuthorization, error) {
	form := url.Values{"client_id": {c.clientID}}
	var raw struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
	if err := c.postForm(ctx, "/user_management/authorize/device", form, &raw); err != nil {
		return DeviceAuthorization{}, err
	}
	interval := time.Duration(raw.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	expires := time.Duration(raw.ExpiresIn) * time.Second
	if expires <= 0 {
		expires = 5 * time.Minute
	}
	return DeviceAuthorization{
		DeviceCode:              raw.DeviceCode,
		UserCode:                raw.UserCode,
		VerificationURI:         raw.VerificationURI,
		VerificationURIComplete: raw.VerificationURIComplete,
		ExpiresIn:               expires,
		Interval:                interval,
	}, nil
}

func (c *Client) PollDevice(ctx context.Context, auth DeviceAuthorization) (Tokens, error) {
	deadline := time.Now().Add(auth.ExpiresIn)
	interval := auth.Interval
	form := url.Values{
		"grant_type":  {deviceCodeGrant},
		"device_code": {auth.DeviceCode},
		"client_id":   {c.clientID},
	}
	for {
		if time.Now().After(deadline) {
			return Tokens{}, fmt.Errorf("device authorization expired")
		}
		tok, pending, err := c.authenticate(ctx, form)
		if err != nil {
			return Tokens{}, err
		}
		if !pending {
			return tok, nil
		}
		select {
		case <-ctx.Done():
			return Tokens{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.clientID},
	}
	tok, pending, err := c.authenticate(ctx, form)
	if err != nil {
		return Tokens{}, err
	}
	if pending {
		return Tokens{}, fmt.Errorf("unexpected pending refresh")
	}
	return tok, nil
}

func (c *Client) authenticate(ctx context.Context, form url.Values) (Tokens, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/user_management/authenticate", strings.NewReader(form.Encode()))
	if err != nil {
		return Tokens{}, false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		return Tokens{}, false, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Tokens{}, false, err
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		tok, err := decodeTokens(body)
		return tok, false, err
	}
	var ae authError
	_ = json.Unmarshal(body, &ae)
	switch ae.Error {
	case "authorization_pending":
		return Tokens{}, true, nil
	case "slow_down":
		return Tokens{}, true, nil
	case "access_denied":
		return Tokens{}, false, fmt.Errorf("device authorization denied")
	case "expired_token":
		return Tokens{}, false, fmt.Errorf("device authorization expired")
	default:
		return Tokens{}, false, fmt.Errorf("authenticate failed: %s (%s)", res.Status, strings.TrimSpace(string(body)))
	}
}

func (c *Client) postForm(ctx context.Context, path string, form url.Values, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s failed: %s (%s)", path, res.Status, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}

func decodeTokens(body []byte) (Tokens, error) {
	var raw struct {
		AccessToken    string `json:"access_token"`
		RefreshToken   string `json:"refresh_token"`
		OrganizationID string `json:"organization_id"`
		ExpiresIn      int    `json:"expires_in"`
		User           struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Tokens{}, err
	}
	if raw.AccessToken == "" {
		return Tokens{}, fmt.Errorf("missing access_token")
	}
	expires := time.Now().Add(time.Hour)
	if raw.ExpiresIn > 0 {
		expires = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)
	}
	return Tokens{
		AccessToken:    raw.AccessToken,
		RefreshToken:   raw.RefreshToken,
		OrganizationID: raw.OrganizationID,
		UserID:         raw.User.ID,
		UserEmail:      raw.User.Email,
		ExpiresAt:      expires,
	}, nil
}
