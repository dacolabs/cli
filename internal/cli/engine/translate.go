// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
	"github.com/dacolabs/daco/internal/translate"
	"github.com/dacolabs/daco/internal/translate/registry"
)

type SchemasTranslateInput struct {
	Prj       *settings.Project
	Schema    string
	Format    string
	OutputDir string
}

type SchemasTranslateOutput struct {
	Schema     string
	Format     string
	OutputFile string
}

func (in SchemasTranslateInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Schema == "" {
		problems["schema"] = "required"
	}
	if in.Format == "" {
		problems["format"] = "required"
	}
	if in.OutputDir == "" {
		problems["output-dir"] = "required"
	}
	return problems
}

func SchemasTranslate(ctx context.Context, in SchemasTranslateInput) (*SchemasTranslateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	path, ok := in.Prj.Schemas[in.Schema]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", in.Schema)
	}
	abs := resolveProjectPath(in.Prj, path)
	s, err := opendpi.LoadSchema(abs)
	if err != nil {
		return nil, fmt.Errorf("load schema %s: %w", abs, err)
	}

	out, err := runTranslator(in.Schema, &s.Schema, in.Format, in.OutputDir)
	if err != nil {
		return nil, err
	}
	return &SchemasTranslateOutput{Schema: in.Schema, Format: in.Format, OutputFile: out}, nil
}

type PortsTranslateInput struct {
	Prj       *settings.Project
	Product   string
	Port      string
	Format    string
	OutputDir string
}

type PortsTranslateOutput struct {
	Product    string
	Port       string
	Format     string
	OutputFile string
}

func (in PortsTranslateInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Port == "" {
		problems["port"] = "required"
	}
	if in.Format == "" {
		problems["format"] = "required"
	}
	if in.OutputDir == "" {
		problems["output-dir"] = "required"
	}
	return problems
}

func PortsTranslate(ctx context.Context, in PortsTranslateInput) (*PortsTranslateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	port, ok := doc.Ports[in.Port]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Port, in.Product)
	}
	if port.Schema == nil {
		return nil, fmt.Errorf("port %q has no schema", in.Port)
	}

	var resolved *jsonschema.Schema
	if port.Schema.Ref != "" {
		schemaPath := port.Schema.Ref
		if !filepath.IsAbs(schemaPath) {
			schemaPath = filepath.Join(filepath.Dir(productPath), schemaPath)
		}
		s, err := opendpi.LoadSchema(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("load referenced schema %s: %w", schemaPath, err)
		}
		resolved = &s.Schema
	} else {
		resolved = &port.Schema.Schema
	}

	out, err := runTranslator(in.Port, resolved, in.Format, in.OutputDir)
	if err != nil {
		return nil, err
	}
	return &PortsTranslateOutput{Product: in.Product, Port: in.Port, Format: in.Format, OutputFile: out}, nil
}

func runTranslator(name string, schema *jsonschema.Schema, format, outputDir string) (string, error) {
	reg := registry.Default()
	t, err := reg.Get(format)
	if err != nil {
		available := reg.Available()
		sort.Strings(available)
		return "", fmt.Errorf("unknown format %q. Available: %v", format, available)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	data, err := t.Translate(name, schema, outputDir)
	if err != nil {
		return "", fmt.Errorf("translate: %w", err)
	}
	if len(data) == 0 {
		return "", errors.New("translator returned empty output")
	}
	outFile := filepath.Join(outputDir, name+ext(t))
	if err := os.WriteFile(outFile, data, 0o644); err != nil {
		return "", err
	}
	return outFile, nil
}

func ext(t translate.Translator) string {
	e := t.FileExtension()
	if e == "" {
		return ""
	}
	if e[0] != '.' {
		return "." + e
	}
	return e
}