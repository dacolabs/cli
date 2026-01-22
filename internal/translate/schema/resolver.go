// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package schema provides shared utilities for JSON Schema processing.
package schema

import (
	"sort"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// Resolver handles $ref resolution and property ordering within a JSON Schema.
type Resolver struct {
	defs      map[string]*jsonschema.Schema
	defNames  []string            // Sorted definition names
	generated map[string]bool     // Track which definitions have been generated
	keyOrder  map[string][]string // Maps JSON path to ordered property names
}

// NewResolver creates a new Resolver for the given schema.
func NewResolver(schema *jsonschema.Schema, keyOrder map[string][]string) *Resolver {
	defs := make(map[string]*jsonschema.Schema)
	for name, def := range schema.Defs {
		defs[name] = def
	}
	for name, def := range schema.Definitions {
		defs[name] = def
	}

	defNames := make([]string, 0, len(defs))
	for name := range defs {
		defNames = append(defNames, name)
	}
	sort.Strings(defNames)

	return &Resolver{
		defs:      defs,
		defNames:  defNames,
		generated: make(map[string]bool),
		keyOrder:  keyOrder,
	}
}

// GetDefNames returns the sorted list of definition names.
func (r *Resolver) GetDefNames() []string {
	return r.defNames
}

// GetDef returns the schema for a definition by name.
func (r *Resolver) GetDef(name string) *jsonschema.Schema {
	return r.defs[name]
}

// MarkGenerated marks a definition as generated.
func (r *Resolver) MarkGenerated(name string) {
	r.generated[name] = true
}

// IsGenerated returns whether a definition has been generated.
func (r *Resolver) IsGenerated(name string) bool {
	return r.generated[name]
}

// GetPropertyOrder returns property names in their original JSON order if available,
// otherwise returns them sorted alphabetically for deterministic output.
func (r *Resolver) GetPropertyOrder(path string, props map[string]*jsonschema.Schema) []string {
	order, ok := r.keyOrder[path]
	if !ok {
		return SortedKeys(props)
	}

	seen := make(map[string]bool, len(props))
	result := make([]string, 0, len(props))
	for _, key := range order {
		if _, exists := props[key]; exists {
			result = append(result, key)
			seen[key] = true
		}
	}
	for key := range props {
		if !seen[key] {
			result = append(result, key)
		}
	}
	return result
}

// GetRefDefName extracts the definition name from a $ref string.
func (r *Resolver) GetRefDefName(ref string) string {
	if !strings.HasPrefix(ref, "#/") {
		return ""
	}
	path := strings.TrimPrefix(ref, "#/")
	switch {
	case strings.HasPrefix(path, "$defs/"):
		return strings.TrimPrefix(path, "$defs/")
	case strings.HasPrefix(path, "definitions/"):
		return strings.TrimPrefix(path, "definitions/")
	case strings.HasPrefix(path, "components/schemas/"):
		return strings.TrimPrefix(path, "components/schemas/")
	}
	return ""
}

// ResolveRef resolves a $ref to its schema.
func (r *Resolver) ResolveRef(ref string) *jsonschema.Schema {
	name := r.GetRefDefName(ref)
	if name == "" {
		return nil
	}
	return r.defs[name]
}

// CollectUsedDefs collects all definitions transitively used by a schema.
func (r *Resolver) CollectUsedDefs(schema *jsonschema.Schema) map[string]bool {
	used := make(map[string]bool)
	var collect func(s *jsonschema.Schema)
	collect = func(s *jsonschema.Schema) {
		if s == nil {
			return
		}
		if s.Ref != "" {
			defName := r.GetRefDefName(s.Ref)
			if defName != "" && !used[defName] {
				used[defName] = true
				if def := r.defs[defName]; def != nil {
					collect(def)
				}
			}
			return
		}
		for _, prop := range s.Properties {
			collect(prop)
		}
		if s.Items != nil {
			collect(s.Items)
		}
		for _, sub := range s.AllOf {
			collect(sub)
		}
	}
	collect(schema)
	return used
}

// CollectDependencies returns the names of all $defs directly referenced by a schema.
func (r *Resolver) CollectDependencies(schema *jsonschema.Schema) []string {
	if schema == nil {
		return nil
	}
	var deps []string
	seen := make(map[string]bool)
	var collect func(s *jsonschema.Schema)
	collect = func(s *jsonschema.Schema) {
		if s == nil {
			return
		}
		if s.Ref != "" {
			name := r.GetRefDefName(s.Ref)
			if name != "" && !seen[name] {
				seen[name] = true
				deps = append(deps, name)
			}
			return
		}
		for _, prop := range s.Properties {
			collect(prop)
		}
		if s.Items != nil {
			collect(s.Items)
		}
		for _, sub := range s.AllOf {
			collect(sub)
		}
	}
	collect(schema)
	return deps
}

// SortedKeys returns map keys sorted alphabetically.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
