// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package opendpi provides OpenDPI specification types.
package opendpi

import (
	"github.com/google/jsonschema-go/jsonschema"
)

// Spec represents the root structure of an OpenDPI specification file.
type Spec struct {
	OpenDPI     string
	Info        Info
	Tags        []Tag
	Connections map[string]Connection
	Ports       map[string]Port
	Schemas     map[string]*jsonschema.Schema // all schemas unified, fully resolved
}

// Info contains metadata about the data product.
type Info struct {
	Title       string
	Version     string
	Description string
}

// Tag is used for categorizing ports.
type Tag struct {
	Name        string
	Description string
}

// Connection represents an infrastructure endpoint where data resides.
type Connection struct {
	Protocol    string
	Host        string
	Description string
	Variables   map[string]any
}

// Port represents a data output interface exposed by the data product.
type Port struct {
	Description string
	Connections []PortConnection
	Schema      *jsonschema.Schema
}

// PortConnection represents a connection-location pair for a port.
type PortConnection struct {
	Connection *Connection
	Location   string
}
