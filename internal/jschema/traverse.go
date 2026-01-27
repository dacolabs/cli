// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import (
	"iter"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// RefResolver resolves $ref strings to schemas.
// Return nil if the ref cannot be resolved.
type RefResolver func(ref string) *jsonschema.Schema

// Traverse returns an iterator over all schemas in the tree.
// It handles cycles by tracking visited schemas.
// If resolver is provided, it follows $ref links to their targets.
func Traverse(schema *jsonschema.Schema, resolver RefResolver) iter.Seq[*jsonschema.Schema] {
	return func(yield func(*jsonschema.Schema) bool) {
		visited := make(map[*jsonschema.Schema]struct{})
		traverseWithVisited(schema, resolver, yield, visited)
	}
}

func traverseWithVisited(schema *jsonschema.Schema, resolver RefResolver, yield func(*jsonschema.Schema) bool, visited map[*jsonschema.Schema]struct{}) bool {
	if schema == nil {
		return true
	}
	if _, ok := visited[schema]; ok {
		return true
	}
	visited[schema] = struct{}{}

	if !yield(schema) {
		return false
	}

	// Follow $ref if resolver is provided
	if schema.Ref != "" && resolver != nil {
		if resolved := resolver(schema.Ref); resolved != nil {
			if !traverseWithVisited(resolved, resolver, yield, visited) {
				return false
			}
		}
	}

	// Objects
	for _, s := range schema.Properties {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.PatternProperties {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !traverseWithVisited(schema.AdditionalProperties, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.PropertyNames, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.UnevaluatedProperties, resolver, yield, visited) {
		return false
	}

	// Arrays
	if !traverseWithVisited(schema.Items, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.ItemsArray {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.PrefixItems {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !traverseWithVisited(schema.AdditionalItems, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.Contains, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.UnevaluatedItems, resolver, yield, visited) {
		return false
	}

	// Logic
	for _, s := range schema.AllOf {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.AnyOf {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.OneOf {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !traverseWithVisited(schema.Not, resolver, yield, visited) {
		return false
	}

	// Conditional
	if !traverseWithVisited(schema.If, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.Then, resolver, yield, visited) {
		return false
	}
	if !traverseWithVisited(schema.Else, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.DependentSchemas {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.DependencySchemas {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}

	// Other
	if !traverseWithVisited(schema.ContentSchema, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.Defs {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.Definitions {
		if !traverseWithVisited(s, resolver, yield, visited) {
			return false
		}
	}

	return true
}

// TraverseDefs returns an iterator over $defs in topological order.
// Dependencies are yielded before the schemas that reference them.
// Each iteration yields (name, schema) pairs.
func TraverseDefs(schema *jsonschema.Schema) iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(string, *jsonschema.Schema) bool) {
		if len(schema.Defs) == 0 {
			return
		}

		order := topologicalSort(schema.Defs)
		for _, name := range order {
			if !yield(name, schema.Defs[name]) {
				return
			}
		}
	}
}

// topologicalSort returns def names ordered so dependencies come first.
// Uses Kahn's algorithm for topological sorting.
func topologicalSort(defs map[string]*jsonschema.Schema) []string {
	// Build adjacency list: for each def, find what other defs it references
	deps := make(map[string][]string)
	for name, s := range defs {
		deps[name] = collectDefRefs(s)
	}

	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for name := range defs {
		inDegree[name] = 0
	}
	for _, refs := range deps {
		for _, ref := range refs {
			if _, exists := defs[ref]; exists {
				inDegree[ref]++
			}
		}
	}

	// Find all nodes with no incoming edges (no dependents)
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Process nodes in order
	var result []string
	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Reduce in-degree for dependencies
		for _, dep := range deps[node] {
			if _, exists := defs[dep]; exists {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					queue = append(queue, dep)
				}
			}
		}
	}

	// Reverse to get dependencies first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// collectDefRefs extracts $ref names from a schema that point to $defs.
func collectDefRefs(schema *jsonschema.Schema) []string {
	var refs []string
	for s := range Traverse(schema, nil) {
		if s.Ref != "" {
			if name, ok := strings.CutPrefix(s.Ref, "#/$defs/"); ok {
				refs = append(refs, name)
			} else if name, ok := strings.CutPrefix(s.Ref, "#/components/schemas/"); ok {
				refs = append(refs, name)
			}
		}
	}
	return refs
}

// RewriteRefs rewrites all $ref attributes to use #/$defs/Name format.
// It handles both #/components/schemas/Name and #/definitions/Name refs.
func RewriteRefs(schema *jsonschema.Schema) {
	for s := range Traverse(schema, nil) {
		if name, ok := strings.CutPrefix(s.Ref, "#/components/schemas/"); ok {
			s.Ref = "#/$defs/" + name
		} else if name, ok := strings.CutPrefix(s.Ref, "#/definitions/"); ok {
			s.Ref = "#/$defs/" + name
		}
	}
}
