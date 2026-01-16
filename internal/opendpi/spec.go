// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package opendpi provides OpenDPI specification types.
package opendpi

// Spec represents the root structure of an OpenDPI specification file.
type Spec struct {
	OpenDPI     string
	Info        Info
	Tags        []Tag
	Connections map[string]Connection
	Ports       map[string]Port
	rawSpec     *rawSpec
}

// NewSpec creates a new spec with provided info and empty collections.
func NewSpec(title, version, description string) *Spec {
	return &Spec{
		OpenDPI: "1.0.0",
		Info: Info{
			Title:       title,
			Version:     version,
			Description: description,
		},
		Connections: make(map[string]Connection),
		Ports:       make(map[string]Port),
	}
}
