package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dacolabs/cli/internal/creds"
	"github.com/dacolabs/cli/internal/session"
)

type memStore struct {
	mu      sync.Mutex
	session creds.Session
	errLoad error
	errSave error
}

func (m *memStore) Load() (creds.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errLoad != nil {
		return creds.Session{}, m.errLoad
	}
	return m.session, nil
}

func (m *memStore) Save(s creds.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errSave != nil {
		return m.errSave
	}
	m.session = s
	return nil
}

func TestAccessTokenUsesValidTokenWithoutRefresh(t *testing.T) {
	store := &memStore{session: creds.Session{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ClientID:     "client",
		ExpiresAt:    time.Now().Add(time.Hour),
	}}
	tok, sess, err := session.AccessToken(context.Background(), store, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "access" || sess.AccessToken != "access" {
		t.Fatalf("tok=%q sess=%+v", tok, sess)
	}
}

func TestAccessTokenRefreshesNearExpiry(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("grant_type") != "refresh_token" || values.Get("refresh_token") != "old_refresh" {
			t.Errorf("form=%s", body)
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":    "new_access",
			"refresh_token":   "new_refresh",
			"organization_id": "org_1",
			"expires_in":      3600,
			"user":            map[string]string{"id": "user_1", "email": "a@example.com"},
		})
	}))
	t.Cleanup(srv.Close)

	store := &memStore{session: creds.Session{
		AccessToken:    "stale",
		RefreshToken:   "old_refresh",
		ClientID:       "client_test",
		OrganizationID: "org_0",
		AuthAPI:        srv.URL,
		ExpiresAt:      time.Now().Add(10 * time.Second), // within 30s skew
	}}
	tok, sess, err := session.AccessToken(context.Background(), store, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
	if tok != "new_access" || sess.RefreshToken != "new_refresh" || sess.OrganizationID != "org_1" {
		t.Fatalf("%+v", sess)
	}
	if store.session.AccessToken != "new_access" {
		t.Fatalf("store not updated: %+v", store.session)
	}
}

func TestAccessTokenSaveFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new_access",
			"refresh_token": "new_refresh",
			"expires_in":    3600,
		})
	}))
	t.Cleanup(srv.Close)

	store := &memStore{
		session: creds.Session{
			AccessToken:  "stale",
			RefreshToken: "old_refresh",
			ClientID:     "client_test",
			AuthAPI:      srv.URL,
			ExpiresAt:    time.Now().Add(-time.Minute),
		},
		errSave: errors.New("disk full"),
	}
	tok, sess, err := session.AccessToken(context.Background(), store, srv.Client())
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("err=%v", err)
	}
	if tok != "new_access" || sess.RefreshToken != "new_refresh" {
		t.Fatalf("expected refreshed tokens despite save failure: tok=%q sess=%+v", tok, sess)
	}
}

func TestAccessTokenExpiredWithoutRefresh(t *testing.T) {
	store := &memStore{session: creds.Session{
		AccessToken:  "stale",
		RefreshToken: "",
		ClientID:     "client",
		ExpiresAt:    time.Now().Add(-time.Minute),
	}}
	_, _, err := session.AccessToken(context.Background(), store, http.DefaultClient)
	if err == nil || err.Error() != "session expired; run daco login" {
		t.Fatalf("err=%v", err)
	}
}
