package creds_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dacolabs/cli/internal/creds"
)

func TestStoreSaveLoadClear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	store := creds.FileStore{Path: path}

	if _, err := store.Load(); err == nil {
		t.Fatal("expected missing credentials error")
	}

	in := creds.Session{
		ClientID:       "client_1",
		AccessToken:    "access",
		RefreshToken:   "refresh",
		OrganizationID: "org_1",
		UserID:         "user_1",
		UserEmail:      "a@example.com",
		ExpiresAt:      time.Now().UTC().Add(time.Hour).Truncate(time.Second),
		AuthAPI:        "https://api.workos.com",
		BaseURL:        "https://main.app.daco.services",
	}
	if err := store.Save(in); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perms := info.Mode().Perm(); perms != 0o600 {
		t.Fatalf("permissions=%o, want 0600", perms)
	}
	out, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != in.AccessToken ||
		out.RefreshToken != in.RefreshToken ||
		out.OrganizationID != in.OrganizationID ||
		out.ClientID != in.ClientID ||
		out.UserID != in.UserID ||
		out.UserEmail != in.UserEmail ||
		out.AuthAPI != in.AuthAPI ||
		out.BaseURL != in.BaseURL ||
		!out.ExpiresAt.Equal(in.ExpiresAt) {
		t.Fatalf("%+v vs %+v", out, in)
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("expected cleared")
	}
}

func TestStaleLockIsRecovered(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	store := creds.FileStore{Path: path}
	lock := path + ".lock"
	if err := os.WriteFile(lock, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-2 * creds.StaleLockAge)
	if err := os.Chtimes(lock, past, past); err != nil {
		t.Fatal(err)
	}
	in := creds.Session{
		ClientID: "c", AccessToken: "a", RefreshToken: "r",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := store.Save(in); err != nil {
		t.Fatal(err)
	}
}
