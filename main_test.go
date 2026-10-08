package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunHelp(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
}

func TestRunVersion(t *testing.T) {
	if err := run([]string{"--version"}); err != nil {
		t.Fatalf("version: %v", err)
	}
}

func TestLoginDefaultsDoNotRequireManualClientID(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_ENV", "staging")
	t.Setenv("DACO_AUTH_API", "http://127.0.0.1:9") // force quick network failure after config resolves
	t.Setenv("DACO_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "creds.json"))
	err := run([]string{"login"})
	if err == nil {
		t.Fatal("expected auth API failure after resolving defaults")
	}
}

func TestLogoutClearsMissingFile(t *testing.T) {
	t.Setenv("DACO_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "missing.json"))
	if err := run([]string{"logout"}); err != nil {
		t.Fatal(err)
	}
}

func TestWhoamiWithoutLogin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	t.Setenv("DACO_CREDENTIALS_FILE", path)
	_ = os.Remove(path)
	if err := run([]string{"whoami"}); err == nil {
		t.Fatal("expected error")
	}
}
