// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import "iter"

// RefResolver resolves $ref strings to schemas.
// Return nil if the ref cannot be resolved.
type RefResolver func(ref string) *Schema

// Traverse returns an iterator over all schemas in the tree.
// It handles cycles by tracking visited schemas.
// If resolver is provided, it follows $ref links to their targets.
func Traverse(schema *Schema, resolver RefResolver) iter.Seq[*Schema] {
	return func(yield func(*Schema) bool) {
		visited := make(map[*Schema]struct{})
		traverseWithVisited(schema, resolver, yield, visited)
	}
}

func traverseWithVisited(schema *Schema, resolver RefResolver, yield func(*Schema) bool, visited map[*Schema]struct{}) bool {
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
