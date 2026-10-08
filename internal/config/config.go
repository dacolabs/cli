package config

import (
	"fmt"
	"os"
	"strings"
)

const DefaultAuthAPI = "https://api.workos.com"

type Config struct {
	ClientID string
	AuthAPI  string
	BaseURL  string
}

func Load() (Config, error) {
	c := Config{
		ClientID: strings.TrimSpace(first(os.Getenv("DACO_CLIENT_ID"), os.Getenv("WORKOS_CLIENT_ID"))),
		AuthAPI:  strings.TrimSpace(first(os.Getenv("DACO_AUTH_API"), DefaultAuthAPI)),
		BaseURL:  strings.TrimRight(strings.TrimSpace(os.Getenv("DACO_BASE_URL")), "/"),
	}
	if c.AuthAPI == "" {
		c.AuthAPI = DefaultAuthAPI
	}
	return c, nil
}

func (c Config) RequireClientID() error {
	if c.ClientID == "" {
		return fmt.Errorf("set DACO_CLIENT_ID to your WorkOS AuthKit client ID")
	}
	return nil
}

func (c Config) RequireBaseURL() error {
	if c.BaseURL == "" {
		return fmt.Errorf("set DACO_BASE_URL to your Catalog API origin (e.g. https://api.example.com)")
	}
	return nil
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
