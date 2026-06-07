// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package cli

import (
	"context"

	"github.com/dacolabs/daco/internal/cli/settings"
)

type ctxKey string

const (
	userKey    ctxKey = "user"
	projectKey ctxKey = "project"
)

func WithUser(ctx context.Context, u *settings.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func WithProject(ctx context.Context, p *settings.Project) context.Context {
	return context.WithValue(ctx, projectKey, p)
}

func User(ctx context.Context) *settings.User {
	u, _ := ctx.Value(userKey).(*settings.User)
	return u
}

func Project(ctx context.Context) *settings.Project {
	p, _ := ctx.Value(projectKey).(*settings.Project)
	return p
}