// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

// Tag is used for categorizing ports.
type Tag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}
