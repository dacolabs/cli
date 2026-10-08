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

func TestLoginRequiresClientID(t *testing.T) {
	t.Setenv("DACO_CLIENT_ID", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("DACO_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "creds.json"))
	err := run([]string{"login"})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected client id error, got %v", err)
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
