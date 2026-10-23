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

// WithLock serializes credential read/write across processes for the duration of fn.
func (s FileStore) WithLock(fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	lockPath := s.Path + ".lock"
	var f *os.File
	var err error
	for i := 0; i < 100; i++ {
		f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return err
		}
		time.Sleep(20 * time.Millisecond)
	}
	if f == nil {
		return fmt.Errorf("could not acquire credentials lock at %s", lockPath)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(lockPath)
	}()
	return fn()
}

func (s FileStore) Save(session Session) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".credentials-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.Path); err != nil {
		return err
	}
	cleanup = false
	return nil
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
