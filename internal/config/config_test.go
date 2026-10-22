package config_test

import (
	"testing"

	"github.com/dacolabs/cli/internal/config"
)

func TestLoadDefaultsToStagingPublicBinding(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.RequireClientID(); err != nil {
		t.Fatal(err)
	}
	if c.ClientID != config.StagingClientID {
		t.Fatalf("client=%q", c.ClientID)
	}
	if err := c.RequireBaseURL(); err != nil {
		t.Fatal(err)
	}
	if c.BaseURL != config.StagingBaseURL {
		t.Fatalf("base=%q", c.BaseURL)
	}
}

func TestEnvOverridesDefaults(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "client_override")
	t.Setenv("DACO_BASE_URL", "http://127.0.0.1:8080")
	t.Setenv("DACO_ENV", "staging")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "client_override" || c.BaseURL != "http://127.0.0.1:8080" {
		t.Fatalf("%+v", c)
	}
}

func TestProductionRequiresConfiguredBinding(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "production")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.RequireClientID(); err == nil {
		t.Fatal("expected missing production client id")
	}
}
