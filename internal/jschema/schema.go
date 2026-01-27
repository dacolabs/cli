// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package jschema provides JSON Schema loading, parsing, and traversal utilities.
package jschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dacolabs/jsonschema-go/jsonschema"
)

// IsFileRef returns true if ref is an external file reference.
// File refs do not start with "#/".
func IsFileRef(ref string) bool {
	return ref != "" && !strings.HasPrefix(ref, "#/")
}

// ExtractKeyOrder parses raw JSON and extracts the order of keys for all "properties" objects.
func ExtractKeyOrder(s *jsonschema.Schema) (map[string][]string, error) {
	rawJSON, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize schema: %w", err)
	}

	result := make(map[string][]string)
	var extract func(dec *json.Decoder, path string)
	extract = func(dec *json.Decoder, path string) {
		token, err := dec.Token()
		if err != nil {
			return
		}
		if t, ok := token.(json.Delim); ok {
			if t == '{' {
				var keys []string
				for dec.More() {
					keyToken, err := dec.Token()
					if err != nil {
						return
					}
					key, ok := keyToken.(string)
					if !ok {
						continue
					}
					keys = append(keys, key)
					var newPath string
					if path == "" {
						newPath = key
					} else {
						newPath = path + "." + key
					}
					extract(dec, newPath)
				}
				_, _ = dec.Token()
				if strings.HasSuffix(path, "properties") || path == "properties" {
					result[path] = keys
				}
			} else if t == '[' {
				for dec.More() {
					extract(dec, path)
				}
				_, _ = dec.Token()
			}
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(rawJSON)))
	extract(dec, "")
	return result, nil
}
