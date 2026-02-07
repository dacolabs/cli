// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

func loadAndExtractOrder(data []byte, filePath string) (*jsonschema.Schema, map[string][]string, error) {
	var schema jsonschema.Schema

	switch {
	case strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml"):
		if err := yaml.Unmarshal(data, &schema); err != nil {
			return nil, nil, err
		}
		keyOrder, err := ExtractKeyOrderFromYAML(data)
		if err != nil {
			return nil, nil, err
		}
		return &schema, keyOrder, nil
	case strings.HasSuffix(filePath, ".json"):
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, nil, err
		}
		keyOrder, err := ExtractKeyOrderFromJSON(data)
		if err != nil {
			return nil, nil, err
		}
		return &schema, keyOrder, nil
	default:
		return nil, nil, fmt.Errorf("format not supported")
	}
}

// Loader loads schemas from a filesystem.
type Loader struct {
	fsys fs.FS
}

// NewLoader creates a Loader that reads from the given filesystem.
func NewLoader(fsys fs.FS) *Loader {
	return &Loader{fsys: fsys}
}

// LoadFile loads and parses a schema file.
// The format is determined from the file extension.
func (l *Loader) LoadFile(filePath string) (*jsonschema.Schema, error) {
	f, err := l.fsys.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	schema, keyOrder, err := loadAndExtractOrder(data, filePath)
	if err != nil {
		return nil, err
	}

	SetPropertyOrder(schema, keyOrder)

	return schema, nil
}

// ResolveRefs resolves all external file $refs in the schema tree in-place.
// It recursively loads referenced schemas and replaces the ref with the loaded content.
// Internal refs (starting with #/) are left unchanged.
func (l *Loader) ResolveRefs(schema *jsonschema.Schema, basePath string) error {
	for s := range Traverse(schema, nil) {
		if !IsFileRef(s.Ref) {
			continue
		}
		refPath := path.Join(basePath, s.Ref)
		loaded, err := l.LoadFile(refPath)
		if err != nil {
			return err
		}
		newBase := path.Dir(refPath)
		if err := l.ResolveRefs(loaded, newBase); err != nil {
			return err
		}
		*s = *loaded
	}
	return nil
}
