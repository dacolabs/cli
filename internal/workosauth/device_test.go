package workosauth_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dacolabs/cli/internal/workosauth"
)

func TestAuthorizeDevice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_management/authorize/device" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("client_id") != "client_test" {
			t.Fatalf("client_id=%q", values.Get("client_id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":               "device_abc",
			"user_code":                 "ABCD-EFGH",
			"verification_uri":          "https://auth.example/device",
			"verification_uri_complete": "https://auth.example/device?user_code=ABCD-EFGH",
			"expires_in":                300,
			"interval":                  1,
		})
	}))
	t.Cleanup(srv.Close)

	client := workosauth.New(srv.URL, "client_test", srv.Client())
	auth, err := client.AuthorizeDevice(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if auth.DeviceCode != "device_abc" || auth.UserCode != "ABCD-EFGH" || auth.Interval != time.Second {
		t.Fatalf("%+v", auth)
	}
}

func TestPollDeviceSuccess(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_management/authenticate" {
			t.Fatalf("path %s", r.URL.Path)
		}
		calls++
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("grant_type") != "urn:ietf:params:oauth:grant-type:device_code" {
			t.Fatalf("grant=%q", values.Get("grant_type"))
		}
		if calls == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":    "access_tok",
			"refresh_token":   "refresh_tok",
			"organization_id": "org_1",
			"expires_in":      3600,
			"user":            map[string]string{"id": "user_1", "email": "a@example.com"},
		})
	}))
	t.Cleanup(srv.Close)

	client := workosauth.New(srv.URL, "client_test", srv.Client())
	tok, err := client.PollDevice(t.Context(), workosauth.DeviceAuthorization{
		DeviceCode: "device_abc",
		Interval:   time.Millisecond,
		ExpiresIn:  time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "access_tok" || tok.RefreshToken != "refresh_tok" || tok.OrganizationID != "org_1" || tok.UserID != "user_1" {
		t.Fatalf("%+v", tok)
	}
	if calls < 2 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestPollDeviceDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "access_denied"})
	}))
	t.Cleanup(srv.Close)

	client := workosauth.New(srv.URL, "client_test", srv.Client())
	_, err := client.PollDevice(t.Context(), workosauth.DeviceAuthorization{
		DeviceCode: "x",
		Interval:   time.Millisecond,
		ExpiresIn:  time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("err=%v", err)
	}
}

func TestRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("grant_type") != "refresh_token" || values.Get("refresh_token") != "old_refresh" {
			t.Fatalf("%s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":    "new_access",
			"refresh_token":   "new_refresh",
			"organization_id": "org_1",
			"expires_in":      1800,
			"user":            map[string]string{"id": "user_1"},
		})
	}))
	t.Cleanup(srv.Close)

	client := workosauth.New(srv.URL, "client_test", srv.Client())
	tok, err := client.Refresh(t.Context(), "old_refresh")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "new_access" || tok.RefreshToken != "new_refresh" {
		t.Fatalf("%+v", tok)
	}
}
