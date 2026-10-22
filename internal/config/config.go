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
	Env      string
	ClientID string
	AuthAPI  string
	BaseURL  string
}

func Load() (Config, error) {
	env := strings.ToLower(strings.TrimSpace(first(os.Getenv("DACO_ENV"), "staging")))
	defaults := environmentDefaults(env)

	c := Config{
		Env:      env,
		ClientID: strings.TrimSpace(first(os.Getenv("DACO_CLIENT_ID"), os.Getenv("WORKOS_CLIENT_ID"), defaults.ClientID)),
		AuthAPI:  strings.TrimSpace(first(os.Getenv("DACO_AUTH_API"), DefaultAuthAPI)),
		BaseURL:  strings.TrimRight(strings.TrimSpace(first(os.Getenv("DACO_BASE_URL"), defaults.BaseURL)), "/"),
	}
	if c.AuthAPI == "" {
		c.AuthAPI = DefaultAuthAPI
	}
	return c, nil
}

func (c Config) RequireClientID() error {
	if c.ClientID == "" {
		if c.Env == "production" {
			return fmt.Errorf("production AuthKit client ID is not baked into this CLI yet; set DACO_CLIENT_ID or use DACO_ENV=staging")
		}
		return fmt.Errorf("set DACO_CLIENT_ID or DACO_ENV to a supported environment")
	}
	return nil
}

func (c Config) RequireBaseURL() error {
	if c.BaseURL == "" {
		return fmt.Errorf("set DACO_BASE_URL to your Catalog API origin")
	}
	return nil
}

type envDefaults struct {
	ClientID string
	BaseURL  string
}

func environmentDefaults(env string) envDefaults {
	switch env {
	case "production", "prod":
		return envDefaults{ClientID: ProductionClientID, BaseURL: ProductionBaseURL}
	default:
		return envDefaults{ClientID: StagingClientID, BaseURL: StagingBaseURL}
	}
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
