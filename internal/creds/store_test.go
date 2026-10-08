package creds_test

import (
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
	}
	if err := store.Save(in); err != nil {
		t.Fatal(err)
	}
	out, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != in.AccessToken || out.RefreshToken != in.RefreshToken || out.OrganizationID != in.OrganizationID || out.ClientID != in.ClientID {
		t.Fatalf("%+v vs %+v", out, in)
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("expected cleared")
	}
}
