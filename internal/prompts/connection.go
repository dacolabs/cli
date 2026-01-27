// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// ConnectionAddResult holds the result of connection form functions.
type ConnectionAddResult struct {
	Name       string
	Connection opendpi.Connection
}

// RunAddNewConnectionForm runs the interactive form for creating a new connection.
func RunAddNewConnectionForm(existingConns map[string]opendpi.Connection) (ConnectionAddResult, error) {
	var result ConnectionAddResult
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Connection name").
				Placeholder("e.g., kafka_prod, postgres_main").
				Value(&result.Name).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("name is required")
					}
					for i, r := range s {
						if i == 0 && !unicode.IsLetter(r) && r != '_' {
							return errors.New("must start with letter or underscore")
						}
						if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
							return errors.New("must contain only letters, numbers, underscores")
						}
					}
					if _, exists := existingConns[s]; exists {
						return fmt.Errorf("connection %q already exists", s)
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Protocol").
				Options(
					huh.NewOption("kafka", "kafka"),
					huh.NewOption("postgresql", "postgresql"),
					huh.NewOption("mysql", "mysql"),
					huh.NewOption("s3", "s3"),
					huh.NewOption("http", "http"),
					huh.NewOption("grpc", "grpc"),
					huh.NewOption("redis", "redis"),
					huh.NewOption("mongodb", "mongodb"),
					huh.NewOption("other", "other"),
				).
				Value(&result.Connection.Protocol),
			huh.NewInput().
				Title("Host").
				Placeholder("e.g., localhost:9092, db.example.com:5432").
				Value(&result.Connection.Host).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("host is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description (optional)").
				Placeholder("e.g., Production Kafka cluster").
				Value(&result.Connection.Description),
		),
	).Run(); err != nil {
		return ConnectionAddResult{}, err
	}

	return result, nil
}

// RunSelectConnectionForm prompts for selecting one existing connection.
func RunSelectConnectionForm(existingConns map[string]opendpi.Connection) (ConnectionAddResult, error) {
	var result ConnectionAddResult
	options := make([]huh.Option[string], 0, len(existingConns))
	for n, c := range existingConns {
		label := fmt.Sprintf("%s (%s://%s)", n, c.Protocol, c.Host)
		options = append(options, huh.NewOption(label, n))
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select connection").
				Options(options...).
				Filtering(true).
				Value(&result.Name).
				Height(8),
		),
	).Run(); err != nil {
		return ConnectionAddResult{}, err
	}

	result.Connection = existingConns[result.Name]
	return result, nil
}
