package catalogapi

import (
	"context"
	"fmt"
	"net/http"
)

// NewAuthenticatedClient returns a Catalog ClientWithResponses that attaches
// the given Bearer access token to every request.
func NewAuthenticatedClient(baseURL, accessToken string, httpClient *http.Client) (*ClientWithResponses, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing access token")
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
