// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

// Connection represents an infrastructure endpoint where data resides.
type Connection struct {
	Protocol    string
	Host        string
	Description string
	Variables   map[string]any
}
