// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

// Connection represents an infrastructure endpoint where data resides.
type Connection struct {
	Protocol    string         `yaml:"protocol" json:"protocol"`
	Host        string         `yaml:"host" json:"host"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
}
