package creds

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrNotFound = errors.New("no stored credentials; run daco login")

type Session struct {
	ClientID       string    `json:"clientId"`
	AccessToken    string    `json:"accessToken"`
	RefreshToken   string    `json:"refreshToken"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	UserEmail      string    `json:"userEmail"`
	ExpiresAt      time.Time `json:"expiresAt"`
	AuthAPI        string    `json:"authApi"`
	BaseURL        string    `json:"baseUrl,omitempty"`
}

type FileStore struct {
	Path string
}

func DefaultPath() (string, error) {
	if p := os.Getenv("DACO_CREDENTIALS_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "daco", "credentials.json"), nil
}

func (s FileStore) Save(session Session) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.Path)
}

func (s FileStore) Load() (Session, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("corrupt credentials file: %w", err)
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (s FileStore) Clear() error {
	err := os.Remove(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
