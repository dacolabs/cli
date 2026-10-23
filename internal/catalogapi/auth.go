package catalogapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewAuthenticatedClient returns a Catalog ClientWithResponses that attaches
// the given Bearer access token to every request.
func NewAuthenticatedClient(baseURL, accessToken string, httpClient *http.Client) (*ClientWithResponses, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}
	if err := requireSafeBaseURL(baseURL); err != nil {
		return nil, err
	}
	httpClient = SecureHTTPClient(httpClient)
	opts := []ClientOption{
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+accessToken)
			return nil
		}),
		WithHTTPClient(httpClient),
	}
	return NewClientWithResponses(baseURL, opts...)
}

// SecureHTTPClient returns a client that refuses redirects to non-https origins
// (except loopback http), so Bearer tokens are not sent in cleartext after a redirect.
func SecureHTTPClient(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	cloned := *httpClient
	prev := cloned.CheckRedirect
	cloned.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := requireSafeRedirect(req.URL); err != nil {
			return err
		}
		if prev != nil {
			return prev(req, via)
		}
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	return &cloned
}

func requireSafeBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid DACO_BASE_URL: %w", err)
	}
	return requireSafeRedirect(u)
}

func requireSafeRedirect(u *url.URL) error {
	if u == nil {
		return fmt.Errorf("missing URL")
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(u.Hostname()) {
			return nil
		}
		return fmt.Errorf("refusing cleartext URL %s; http is only allowed for localhost", u.String())
	default:
		return fmt.Errorf("refusing non-https URL %s", u.String())
	}
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
