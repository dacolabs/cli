// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrors_Error(t *testing.T) {
	e := Errors{"name: required", "path: required"}
	assert.Equal(t, "name: required; path: required", e.Error())
}

func TestErrors_AsRecovers(t *testing.T) {
	var err error = Errors{"a", "b"}
	var list Errors
	require.True(t, errors.As(err, &list))
	assert.Equal(t, Errors{"a", "b"}, list)
}

type validInput struct {
	problems map[string]string
}

func (v validInput) Valid(ctx context.Context) map[string]string {
	return v.problems
}

func TestValidate_Empty(t *testing.T) {
	assert.NoError(t, validate(context.Background(), validInput{}))
	assert.NoError(t, validate(context.Background(), validInput{problems: map[string]string{}}))
}

func TestValidate_Sorted(t *testing.T) {
	err := validate(context.Background(), validInput{problems: map[string]string{
		"zebra": "required",
		"apple": "required",
		"mango": "required",
	}})
	require.Error(t, err)
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Equal(t, Errors{"apple: required", "mango: required", "zebra: required"}, errs)
}