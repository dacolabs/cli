package catalogapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	opts := []ClientOption{
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+accessToken)
			return nil
		}),
	}
	if httpClient != nil {
		opts = append(opts, WithHTTPClient(httpClient))
	}
	return NewClientWithResponses(baseURL, opts...)
}

func requireSafeBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid DACO_BASE_URL: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if isLoopbackHost(host) {
			return nil
		}
		return fmt.Errorf("DACO_BASE_URL must use https (got %s); http is only allowed for localhost", baseURL)
	default:
		return fmt.Errorf("DACO_BASE_URL must use https (got %s)", baseURL)
	}
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
