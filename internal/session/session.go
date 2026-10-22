package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dacolabs/cli/internal/creds"
	"github.com/dacolabs/cli/internal/workosauth"
)

type Store interface {
	Load() (creds.Session, error)
	Save(creds.Session) error
}

func AccessToken(ctx context.Context, store Store, httpClient *http.Client) (string, creds.Session, error) {
	session, err := store.Load()
	if err != nil {
		return "", creds.Session{}, err
	}
	if time.Now().Before(session.ExpiresAt.Add(-30 * time.Second)) {
		return session.AccessToken, session, nil
	}
	if session.RefreshToken == "" || session.ClientID == "" {
		return "", creds.Session{}, fmt.Errorf("session expired; run daco login")
	}
	authAPI := session.AuthAPI
	if authAPI == "" {
		authAPI = "https://api.workos.com"
	}
	client := workosauth.New(authAPI, session.ClientID, httpClient)
	tok, err := client.Refresh(ctx, session.RefreshToken)
	if err != nil {
		return "", creds.Session{}, fmt.Errorf("refresh failed (%w); run daco login", err)
	}
	session.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		session.RefreshToken = tok.RefreshToken
	}
	if tok.OrganizationID != "" {
		session.OrganizationID = tok.OrganizationID
	}
	if tok.UserID != "" {
		session.UserID = tok.UserID
	}
	if tok.UserEmail != "" {
		session.UserEmail = tok.UserEmail
	}
	session.ExpiresAt = tok.ExpiresAt
	if err := store.Save(session); err != nil {
		return "", creds.Session{}, err
	}
	return session.AccessToken, session, nil
}
