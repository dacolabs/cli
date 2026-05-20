package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// maxRequestBodySize is the maximum allowed request body size (10MB).
const maxRequestBodySize = 10 * 1024 * 1024

// encode writes the response as JSON with the given status code.
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// decode reads and decodes JSON from the request body.
// It limits the request body size to MaxRequestBodySize to prevent large payload attacks.
func decode[T any](r *http.Request) (T, error) {
	var v T
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// decodeValid reads, decodes, and validates JSON from the request body.
// It returns the decoded value, a map of validation problems, and any error encountered.
func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
    v, err := decode[T](r)
    if err != nil {
		return v, nil, err
	}
    if problems := v.Valid(r.Context()); len(problems) > 0 {
        return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
    }
    return v, nil, nil
}