package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultAuthAPI = "https://api.workos.com"

	// Public AuthKit client IDs (not secrets). Safe to ship in a public CLI.
	StagingClientID = "client_01KC6DJVAFP219V90F6CCQNFYY"
	StagingBaseURL  = "https://main.app.daco.services"

	// ProductionClientID is empty until the production WorkOS binding is checked in.
	ProductionClientID = ""
	ProductionBaseURL  = "https://app.dacolabs.com"
)

type Config struct {
	Env string
	// ClientID is the AuthKit client ID.
	ClientID string
	AuthAPI  string
	// BaseURL is set only when DACO_BASE_URL is present (explicit override).
	BaseURL string
	// DefaultBaseURL comes from DACO_ENV and is used when neither DACO_BASE_URL
	// nor a saved session BaseURL is set.
	DefaultBaseURL string
}

func Load() (Config, error) {
	env := strings.ToLower(strings.TrimSpace(first(os.Getenv("DACO_ENV"), "staging")))
	defaults, err := environmentDefaults(env)
	if err != nil {
		return Config{}, err
	}

	// Prefer DACO_CLIENT_ID, then baked-in defaults, then WORKOS_CLIENT_ID as a last
	// resort (so a shell WORKOS_CLIENT_ID cannot override the staging binding).
	c := Config{
		Env:            env,
		ClientID:       strings.TrimSpace(first(os.Getenv("DACO_CLIENT_ID"), defaults.ClientID, os.Getenv("WORKOS_CLIENT_ID"))),
		AuthAPI:        strings.TrimSpace(first(os.Getenv("DACO_AUTH_API"), DefaultAuthAPI)),
		BaseURL:        strings.TrimRight(strings.TrimSpace(os.Getenv("DACO_BASE_URL")), "/"),
		DefaultBaseURL: strings.TrimRight(defaults.BaseURL, "/"),
	}
	if c.AuthAPI == "" {
		c.AuthAPI = DefaultAuthAPI
	}
	return c, nil
}

// ResolveBaseURL prefers an explicit DACO_BASE_URL, then a saved session URL, then env defaults.
func (c Config) ResolveBaseURL(sessionBase string) string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	if strings.TrimSpace(sessionBase) != "" {
		return strings.TrimRight(strings.TrimSpace(sessionBase), "/")
	}
	return c.DefaultBaseURL
}

func (c Config) RequireClientID() error {
	if c.ClientID == "" {
		if isProductionEnv(c.Env) {
			return fmt.Errorf("production AuthKit client ID is not baked into this CLI yet; set DACO_CLIENT_ID or use DACO_ENV=staging")
		}
		return fmt.Errorf("set DACO_CLIENT_ID or DACO_ENV to a supported environment")
	}
	return nil
}

func (c Config) RequireBaseURL() error {
	if c.ResolveBaseURL("") == "" {
		return fmt.Errorf("set DACO_BASE_URL to your Catalog API origin")
	}
	return nil
}

type envDefaults struct {
	ClientID string
	BaseURL  string
}

func environmentDefaults(env string) (envDefaults, error) {
	switch env {
	case "production", "prod":
		return envDefaults{ClientID: ProductionClientID, BaseURL: ProductionBaseURL}, nil
	case "staging":
		return envDefaults{ClientID: StagingClientID, BaseURL: StagingBaseURL}, nil
	default:
		return envDefaults{}, fmt.Errorf("unsupported DACO_ENV %q (want staging, production, or prod)", env)
	}
}

func isProductionEnv(env string) bool {
	return env == "production" || env == "prod"
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
