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

// RunAddNewConnectionForm runs the interactive form for creating a new connection.
// Returns the connection name and the connection object.
func RunAddNewConnectionForm(existingConns map[string]opendpi.Connection) (name string, conn opendpi.Connection, _ error) {
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Connection name").
				Placeholder("e.g., kafka_prod, postgres_main").
				Value(&name).
				Validate(connectionNameValidator(existingConns)),
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
				Value(&conn.Protocol),
			huh.NewInput().
				Title("Host").
				Placeholder("e.g., localhost:9092, db.example.com:5432").
				Value(&conn.Host).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("host is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description (optional)").
				Placeholder("e.g., Production Kafka cluster").
				Value(&conn.Description),
		),
	).Run(); err != nil {
		return name, conn, err
	}

	return name, conn, nil
}

func connectionNameValidator(existing map[string]opendpi.Connection) func(string) error {
	return func(s string) error {
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
		if _, exists := existing[s]; exists {
			return fmt.Errorf("connection %q already exists", s)
		}
		return nil
	}
}

// RunAddConnectionForm runs the interactive form for adding port connections.
// Allows selecting from existing connections and specifying locations.
// Returns a slice of PortConnection (supports multiple connections).
func RunAddConnectionForm(existingConns map[string]opendpi.Connection) ([]opendpi.PortConnection, error) {
	var conns []opendpi.PortConnection
	for {
		options := make([]huh.Option[string], 0, len(existingConns))
		for name, conn := range existingConns {
			label := fmt.Sprintf("%s (%s://%s)", name, conn.Protocol, conn.Host)
			options = append(options, huh.NewOption(label, name))
		}

		var connName, location string
		var addAnother bool

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select connection").
					Options(options...).
					Filtering(true).
					Value(&connName).
					Height(8),
				huh.NewInput().
					Title("Location").
					Placeholder("e.g., events_topic, users_table, /data/events").
					Value(&location).
					Validate(requiredValidator("location")),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add another connection?").
					Affirmative("Yes").
					Negative("No").
					Value(&addAnother),
			),
		).Run(); err != nil {
			return nil, err
		}

		conn := existingConns[connName]
		conns = append(conns, opendpi.PortConnection{
			Connection: &conn,
			Location:   location,
		})

		if !addAnother {
			break
		}
	}
	return conns, nil
}
