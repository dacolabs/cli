// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import "github.com/google/jsonschema-go/jsonschema"

// Components contains reusable definitions for the OpenDPI document.
type Components struct {
	Schemas map[string]*jsonschema.Schema `yaml:"schemas,omitempty" json:"schemas,omitempty"`
}
