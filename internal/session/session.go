package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dacolabs/cli/internal/creds"
	"github.com/dacolabs/cli/internal/workosauth"
)

const refreshTimeout = 30 * time.Second

type Store interface {
	Load() (creds.Session, error)
	Save(creds.Session) error
}

type unlockedSaver interface {
	SaveUnlocked(creds.Session) error
}

type lockStore interface {
	Store
	WithLock(fn func() error) error
}

func AccessToken(ctx context.Context, store Store, httpClient *http.Client) (string, creds.Session, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, refreshTimeout)
		defer cancel()
	}
	if ls, ok := store.(lockStore); ok {
		var token string
		var session creds.Session
		err := ls.WithLock(func() error {
			var err error
			token, session, err = accessTokenUnlocked(ctx, store, httpClient)
			return err
		})
		return token, session, err
	}
	return accessTokenUnlocked(ctx, store, httpClient)
}

func accessTokenUnlocked(ctx context.Context, store Store, httpClient *http.Client) (string, creds.Session, error) {
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
	if err := saveSession(store, session); err != nil {
		// Return refreshed tokens so this process can continue; name the persist failure.
		return session.AccessToken, session, fmt.Errorf("refreshed session but failed to persist credentials (%w)", err)
	}
	return session.AccessToken, session, nil
}

func saveSession(store Store, session creds.Session) error {
	if us, ok := store.(unlockedSaver); ok {
		return us.SaveUnlocked(session)
	}
	return store.Save(session)
}
