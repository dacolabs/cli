// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

// Info contains metadata about the data product.
type Info struct {
	Title       string `yaml:"title" json:"title"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}
