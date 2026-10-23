package workosauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPollWaitCapsAtDeadline(t *testing.T) {
	deadline := time.Now().Add(2 * time.Second)
	got := pollWait(10*time.Second, deadline)
	if got <= 0 || got > 2*time.Second {
		t.Fatalf("wait=%v", got)
	}
	if pollWait(time.Second, time.Now().Add(-time.Second)) != 0 {
		t.Fatal("expected zero when past deadline")
	}
}

func TestPollDeviceSlowDownIncreasesInterval(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "slow_down"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access_tok",
			"refresh_token": "refresh_tok",
			"expires_in":    3600,
		})
	}))
	t.Cleanup(srv.Close)

	var waits []time.Duration
	prev := after
	after = func(d time.Duration) <-chan time.Time {
		waits = append(waits, d)
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { after = prev })

	client := New(srv.URL, "client_test", srv.Client())
	tok, err := client.PollDevice(t.Context(), DeviceAuthorization{
		DeviceCode: "device_abc",
		Interval:   time.Second,
		ExpiresIn:  time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "access_tok" {
		t.Fatalf("%+v", tok)
	}
	if len(waits) != 1 {
		t.Fatalf("waits=%v", waits)
	}
	if waits[0] != 6*time.Second {
		t.Fatalf("want 6s after slow_down (1s+5s), got %v", waits[0])
	}
}
