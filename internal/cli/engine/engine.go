// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type Validator interface {
	Valid(ctx context.Context) map[string]string
}

// Errors carries one or more problem strings produced by the engine. It
// implements the error interface so it composes naturally with normal Go error
// handling; drivers (CLI / TUI) can do errors.As to recover the list and
// render it however they want.
type Errors []string

func (e Errors) Error() string {
	return strings.Join(e, "; ")
}

func validate(ctx context.Context, in Validator) error {
	problems := in.Valid(ctx)
	if len(problems) == 0 {
		return nil
	}
	parts := make([]string, 0, len(problems))
	for k, v := range problems {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}
	slices.Sort(parts)
	return Errors(parts)
}