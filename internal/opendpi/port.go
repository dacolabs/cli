// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import "github.com/google/jsonschema-go/jsonschema"

// Port represents a data output interface exposed by the data product.
type Port struct {
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Connections []PortConnection   `yaml:"connections" json:"connections"`
	Schema      *jsonschema.Schema `yaml:"schema" json:"schema"`
}

// PortConnection represents a connection-location pair for a port.
type PortConnection struct {
	Connection ConnectionRef `yaml:"connection" json:"connection"`
	Location   string        `yaml:"location" json:"location"`
}

// ConnectionRef is a reference to a connection in the connections registry.
type ConnectionRef struct {
	Ref string `yaml:"$ref" json:"$ref"`
}
