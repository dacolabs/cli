// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import "github.com/google/jsonschema-go/jsonschema"

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
