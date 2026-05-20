// Package env provides small helpers for reading typed values out of
// an environment-lookup function. The indirection via `getenv` (rather
// than direct os.Getenv) keeps callers testable without mutating
// process-wide state.
package env

import (
	"strconv"
	"time"
)

// ParseEnv retrieves a string environment variable or returns the default.
func ParseEnv(getenv func(string) string, key, defaultVal string) string {
	if val := getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ParseEnvInt retrieves an integer environment variable or returns the
// default if the variable is unset or not a valid integer.
func ParseEnvInt(getenv func(string) string, key string, defaultVal int) int {
	val := getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// ParseEnvBool retrieves a boolean environment variable or returns the
// default if the variable is unset or not a valid bool.
func ParseEnvBool(getenv func(string) string, key string, defaultVal bool) bool {
	val := getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

// ParseEnvDuration retrieves a duration environment variable or returns
// the default if the variable is unset or not a valid time.Duration.
func ParseEnvDuration(getenv func(string) string, key string, defaultVal time.Duration) time.Duration {
	val := getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}
