// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
)

type ProjectFormatInput struct {
	Prj *settings.Project
}

type ProjectFormatOutput struct {
	Files []string
}

func (in ProjectFormatInput) Valid(ctx context.Context) map[string]string {
	if in.Prj == nil {
		return map[string]string{"prj": "project not loaded"}
	}
	return nil
}

func ProjectFormat(ctx context.Context, in ProjectFormatInput) (*ProjectFormatOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}

	var formatted []string

	for _, name := range sortedKeys(in.Prj.Schemas) {
		abs := resolveProjectPath(in.Prj, in.Prj.Schemas[name])
		s, err := opendpi.LoadSchema(abs)
		if err != nil {
			return nil, fmt.Errorf("format schema %q (%s): %w", name, abs, err)
		}
		if err := opendpi.SaveSchema(abs, s); err != nil {
			return nil, fmt.Errorf("format schema %q (%s): %w", name, abs, err)
		}
		formatted = append(formatted, abs)
	}

	for _, name := range sortedKeys(in.Prj.Connections) {
		abs := resolveProjectPath(in.Prj, in.Prj.Connections[name])
		c, err := opendpi.LoadConnection(abs)
		if err != nil {
			return nil, fmt.Errorf("format connection %q (%s): %w", name, abs, err)
		}
		if err := opendpi.SaveConnection(abs, c); err != nil {
			return nil, fmt.Errorf("format connection %q (%s): %w", name, abs, err)
		}
		formatted = append(formatted, abs)
	}

	for _, name := range sortedKeys(in.Prj.Products) {
		abs := resolveProjectPath(in.Prj, in.Prj.Products[name])
		doc, err := opendpi.Load(abs)
		if err != nil {
			return nil, fmt.Errorf("format product %q (%s): %w", name, abs, err)
		}
		if err := opendpi.Save(abs, doc); err != nil {
			return nil, fmt.Errorf("format product %q (%s): %w", name, abs, err)
		}
		formatted = append(formatted, abs)
	}

	return &ProjectFormatOutput{Files: formatted}, nil
}

type LintIssue struct {
	Path    string
	Message string
}

type ProjectLintInput struct {
	Prj *settings.Project
}

type ProjectLintOutput struct {
	Issues []LintIssue
}

func (in ProjectLintInput) Valid(ctx context.Context) map[string]string {
	if in.Prj == nil {
		return map[string]string{"prj": "project not loaded"}
	}
	return nil
}

func ProjectLint(ctx context.Context, in ProjectLintInput) (*ProjectLintOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}

	var issues []LintIssue
	report := func(path, format string, args ...any) {
		issues = append(issues, LintIssue{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	for _, name := range sortedKeys(in.Prj.Schemas) {
		abs := resolveProjectPath(in.Prj, in.Prj.Schemas[name])
		if _, err := os.Stat(abs); err != nil {
			report(abs, "schema %q: %v", name, err)
			continue
		}
		if _, err := opendpi.LoadSchema(abs); err != nil {
			report(abs, "schema %q: parse error: %v", name, err)
		}
	}

	for _, name := range sortedKeys(in.Prj.Connections) {
		abs := resolveProjectPath(in.Prj, in.Prj.Connections[name])
		if _, err := os.Stat(abs); err != nil {
			report(abs, "connection %q: %v", name, err)
			continue
		}
		conn, err := opendpi.LoadConnection(abs)
		if err != nil {
			report(abs, "connection %q: parse error: %v", name, err)
			continue
		}
		if conn.IsRef() {
			refPath := conn.Ref
			if !filepath.IsAbs(refPath) {
				refPath = filepath.Join(filepath.Dir(abs), refPath)
			}
			if _, err := os.Stat(refPath); err != nil {
				report(abs, "connection %q: $ref %s does not resolve: %v", name, conn.Ref, err)
			}
		} else if conn.Type == "" || conn.Host == "" {
			report(abs, "connection %q: missing type or host", name)
		}
	}

	for _, name := range sortedKeys(in.Prj.Products) {
		abs := resolveProjectPath(in.Prj, in.Prj.Products[name])
		if _, err := os.Stat(abs); err != nil {
			report(abs, "product %q: %v", name, err)
			continue
		}
		doc, err := opendpi.Load(abs)
		if err != nil {
			report(abs, "product %q: parse error: %v", name, err)
			continue
		}
		lintProduct(abs, name, doc, report)
	}

	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		return issues[i].Message < issues[j].Message
	})
	return &ProjectLintOutput{Issues: issues}, nil
}

func lintProduct(productPath, productName string, doc *opendpi.Document, report func(string, string, ...any)) {
	dir := filepath.Dir(productPath)

	for _, connName := range sortedKeys(doc.Connections) {
		conn := doc.Connections[connName]
		if !conn.IsRef() {
			if conn.Type == "" || conn.Host == "" {
				report(productPath, "product %q: connection %q has no $ref and is missing type/host", productName, connName)
			}
			continue
		}
		refPath := conn.Ref
		if !filepath.IsAbs(refPath) {
			refPath = filepath.Join(dir, refPath)
		}
		if _, err := os.Stat(refPath); err != nil {
			report(productPath, "product %q: connection %q $ref %s does not resolve: %v", productName, connName, conn.Ref, err)
		}
	}

	for _, portName := range sortedKeys(doc.Ports) {
		port := doc.Ports[portName]
		if port.Schema != nil && port.Schema.Ref != "" {
			refPath := port.Schema.Ref
			if !filepath.IsAbs(refPath) {
				refPath = filepath.Join(dir, refPath)
			}
			if _, err := os.Stat(refPath); err != nil {
				report(productPath, "product %q: port %q schema $ref %s does not resolve: %v", productName, portName, port.Schema.Ref, err)
			} else if _, err := opendpi.LoadSchema(refPath); err != nil {
				report(productPath, "product %q: port %q schema %s: parse error: %v", productName, portName, port.Schema.Ref, err)
			}
		}
		for _, pc := range port.Connections {
			if pc.Connection == "" {
				report(productPath, "product %q: port %q has a connection binding with empty name", productName, portName)
				continue
			}
			if _, ok := doc.Connections[pc.Connection]; !ok {
				report(productPath, "product %q: port %q references unknown connection %q", productName, portName, pc.Connection)
			}
			if pc.Location == "" {
				report(productPath, "product %q: port %q binding to connection %q is missing location", productName, portName, pc.Connection)
			}
		}
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
