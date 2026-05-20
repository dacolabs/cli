package env

import (
	"testing"
	"time"
)

func envOf(pairs map[string]string) func(string) string {
	return func(k string) string { return pairs[k] }
}

func TestParseEnv(t *testing.T) {
	t.Parallel()
	t.Run("present", func(t *testing.T) {
		got := ParseEnv(envOf(map[string]string{"K": "v"}), "K", "default")
		if got != "v" {
			t.Errorf("got %q, want %q", got, "v")
		}
	})
	t.Run("missing returns default", func(t *testing.T) {
		got := ParseEnv(envOf(map[string]string{}), "K", "default")
		if got != "default" {
			t.Errorf("got %q, want %q", got, "default")
		}
	})
}

func TestParseEnvInt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  string
		want int
	}{
		{"present valid", "42", 42},
		{"present invalid returns default", "not-a-number", 7},
		{"missing returns default", "", 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEnvInt(envOf(map[string]string{"K": tc.val}), "K", 7)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseEnvBool(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  string
		want bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"present invalid returns default", "yes", true},
		{"missing returns default", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEnvBool(envOf(map[string]string{"K": tc.val}), "K", true)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseEnvDuration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  string
		want time.Duration
	}{
		{"valid", "5s", 5 * time.Second},
		{"present invalid returns default", "5seconds", 30 * time.Second},
		{"missing returns default", "", 30 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEnvDuration(envOf(map[string]string{"K": tc.val}), "K", 30*time.Second)
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
