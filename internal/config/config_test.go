package config_test

import (
	"strings"
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
	if c.ResolveBaseURL("") != config.StagingBaseURL {
		t.Fatalf("base=%q", c.ResolveBaseURL(""))
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

func TestSessionBaseURLPreferredOverDefault(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "staging")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := c.ResolveBaseURL("https://custom.example")
	if got != "https://custom.example" {
		t.Fatalf("got %q", got)
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
	reqErr := c.RequireClientID()
	if reqErr == nil {
		t.Fatal("expected missing production client id")
	}
	if !strings.Contains(reqErr.Error(), "production AuthKit client ID") {
		t.Fatalf("err=%v", reqErr)
	}
	if c.DefaultBaseURL != config.ProductionBaseURL {
		t.Fatalf("defaultBase=%q", c.DefaultBaseURL)
	}
}

func TestProdAliasUsesProductionError(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "prod")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	reqErr := c.RequireClientID()
	if reqErr == nil {
		t.Fatal("expected missing production client id")
	}
	if !strings.Contains(reqErr.Error(), "production AuthKit client ID") {
		t.Fatalf("err=%v", reqErr)
	}
}

func TestWorkOSClientIDFallbackWhenNoBakedInDefault(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "client_from_workos")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "production")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "client_from_workos" {
		t.Fatalf("client=%q", c.ClientID)
	}
	if c.DefaultBaseURL != config.ProductionBaseURL {
		t.Fatalf("base=%q", c.DefaultBaseURL)
	}
}

func TestWorkOSClientIDDoesNotOverrideStagingDefault(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "client_from_workos")
	t.Setenv("DACO_BASE_URL", "")
	t.Setenv("DACO_ENV", "staging")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != config.StagingClientID {
		t.Fatalf("client=%q, want baked-in staging", c.ClientID)
	}
}

func TestUnsupportedEnvRejected(t *testing.T) {
	t.Setenv("DACO_ENV", "produciton")
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_BASE_URL", "")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "unsupported DACO_ENV") {
		t.Fatalf("err=%v", err)
	}
}
