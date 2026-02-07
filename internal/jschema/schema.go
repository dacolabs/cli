// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package jschema provides JSON Schema loading, parsing, and traversal utilities.
package jschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
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

// ExtractKeyOrderFromJSON extracts the order of keys for all "properties" objects
// directly from raw JSON bytes, preserving the original file order.
func ExtractKeyOrderFromJSON(data []byte) (map[string][]string, error) {
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
	dec := json.NewDecoder(strings.NewReader(string(data)))
	extract(dec, "")
	return result, nil
}

// ExtractKeyOrderFromYAML extracts the order of keys for all "properties" objects
// from raw YAML bytes, using yaml.Node which preserves insertion order.
func ExtractKeyOrderFromYAML(data []byte) (map[string][]string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	result := make(map[string][]string)
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		ExtractYAMLNodeKeyOrder(root.Content[0], "", result)
	}
	return result, nil
}

// ExtractYAMLNodeKeyOrder recursively extracts key ordering from a yaml.Node tree.
// It collects the ordered keys for all mapping nodes whose path ends with "properties".
func ExtractYAMLNodeKeyOrder(node *yaml.Node, path string, result map[string][]string) {
	if node == nil {
		return
	}

	if node.Kind == yaml.MappingNode {
		var keys []string
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			key := keyNode.Value
			keys = append(keys, key)

			var newPath string
			if path == "" {
				newPath = key
			} else {
				newPath = path + "." + key
			}

			switch valueNode.Kind {
			case yaml.MappingNode:
				ExtractYAMLNodeKeyOrder(valueNode, newPath, result)
			case yaml.SequenceNode:
				for _, item := range valueNode.Content {
					if item.Kind == yaml.MappingNode {
						ExtractYAMLNodeKeyOrder(item, newPath, result)
					}
				}
			}
		}

		if strings.HasSuffix(path, "properties") || path == "properties" {
			result[path] = keys
		}
	}
}

// SetPropertyOrder walks the schema tree and sets PropertyOrder on each node
// that has properties, using the key order extracted from raw bytes.
func SetPropertyOrder(schema *jsonschema.Schema, keyOrder map[string][]string) {
	setPropertyOrderRecursive(schema, "", keyOrder)
}

func setPropertyOrderRecursive(schema *jsonschema.Schema, path string, keyOrder map[string][]string) {
	if schema == nil {
		return
	}

	if len(schema.Properties) > 0 {
		orderPath := "properties"
		if path != "" {
			orderPath = path + ".properties"
		}
		if order, ok := keyOrder[orderPath]; ok {
			var filtered []string
			for _, key := range order {
				if _, exists := schema.Properties[key]; exists {
					filtered = append(filtered, key)
				}
			}
			schema.PropertyOrder = filtered
		}
	}

	for propName, propSchema := range schema.Properties {
		propPath := "properties." + propName
		if path != "" {
			propPath = path + ".properties." + propName
		}
		setPropertyOrderRecursive(propSchema, propPath, keyOrder)
	}

	for defName, defSchema := range schema.Defs {
		defPath := "$defs." + defName
		if path != "" {
			defPath = path + ".$defs." + defName
		}
		setPropertyOrderRecursive(defSchema, defPath, keyOrder)
	}

	for defName, defSchema := range schema.Definitions {
		defPath := "definitions." + defName
		if path != "" {
			defPath = path + ".definitions." + defName
		}
		setPropertyOrderRecursive(defSchema, defPath, keyOrder)
	}

	if schema.Items != nil {
		itemPath := "items"
		if path != "" {
			itemPath = path + ".items"
		}
		setPropertyOrderRecursive(schema.Items, itemPath, keyOrder)
	}

	for _, s := range schema.AllOf {
		setPropertyOrderRecursive(s, path, keyOrder)
	}
	for _, s := range schema.AnyOf {
		setPropertyOrderRecursive(s, path, keyOrder)
	}
	for _, s := range schema.OneOf {
		setPropertyOrderRecursive(s, path, keyOrder)
	}

	if schema.AdditionalProperties != nil {
		apPath := "additionalProperties"
		if path != "" {
			apPath = path + ".additionalProperties"
		}
		setPropertyOrderRecursive(schema.AdditionalProperties, apPath, keyOrder)
	}
}
