// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package schema

import (
	"encoding/json"
	"strings"
)

// ExtractKeyOrder parses raw JSON and extracts the order of keys for all "properties" objects.
func ExtractKeyOrder(rawJSON []byte) map[string][]string {
	result := make(map[string][]string)
	var extract func(dec *json.Decoder, path string)
	extract = func(dec *json.Decoder, path string) {
		token, err := dec.Token()
		if err != nil {
			return
		}
		switch t := token.(type) {
		case json.Delim:
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
				dec.Token() //nolint:errcheck
				if strings.HasSuffix(path, "properties") || path == "properties" {
					result[path] = keys
				}
			} else if t == '[' {
				for dec.More() {
					extract(dec, path)
				}
				dec.Token() //nolint:errcheck
			}
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(rawJSON)))
	extract(dec, "")
	return result
}
