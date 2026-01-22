// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package schema

import (
	"encoding/json"
	"strings"
)

// ExtractKeyOrder parses raw JSON and extracts the order of keys for all "properties" objects.
// Returns a map from JSON path (e.g., "properties", "$defs.address.properties") to ordered keys.
func ExtractKeyOrder(rawJSON []byte) map[string][]string {
	result := make(map[string][]string)

	var extractFromDecoder func(dec *json.Decoder, path string)
	extractFromDecoder = func(dec *json.Decoder, path string) {
		token, err := dec.Token()
		if err != nil {
			return
		}

		switch t := token.(type) {
		case json.Delim:
			if t == '{' {
				// Read object keys in order
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

					// Build the new path
					var newPath string
					if path == "" {
						newPath = key
					} else {
						newPath = path + "." + key
					}

					// Recurse into the value
					extractFromDecoder(dec, newPath)
				}
				// Consume the closing brace
				dec.Token()

				// Store keys if this is a "properties" path
				if strings.HasSuffix(path, "properties") || path == "properties" {
					result[path] = keys
				}
			} else if t == '[' {
				// Read array elements
				for dec.More() {
					extractFromDecoder(dec, path)
				}
				// Consume the closing bracket
				dec.Token()
			}
		}
	}

	dec := json.NewDecoder(strings.NewReader(string(rawJSON)))
	extractFromDecoder(dec, "")

	return result
}
