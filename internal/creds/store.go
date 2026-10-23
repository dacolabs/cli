package creds

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrNotFound = errors.New("no stored credentials; run daco login")

// StaleLockAge is how old a credentials lock may be before it is treated as abandoned.
// Must exceed the session refresh HTTP budget so a live refresh is not interrupted.
const StaleLockAge = 45 * time.Second

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

func (s FileStore) lockPath() string {
	return s.Path + ".lock"
}

func newLockToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// WithLock serializes credential read/write across processes for the duration of fn.
func (s FileStore) WithLock(fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	lockPath := s.lockPath()
	var f *os.File
	var err error
	for i := 0; i < 100; i++ {
		f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			token := newLockToken()
			if _, werr := f.WriteString(token); werr != nil {
				_ = f.Close()
				_ = os.Remove(lockPath)
				return werr
			}
			break
		}
		if !os.IsExist(err) {
			return err
		}
		info, statErr := os.Stat(lockPath)
		if statErr != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		if time.Since(info.ModTime()) <= StaleLockAge {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		staleContent, readErr := os.ReadFile(lockPath)
		if readErr != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		staleMod := info.ModTime()
		// Re-check identity so we never unlink a lock another process just created.
		info2, statErr := os.Stat(lockPath)
		if statErr != nil || !info2.ModTime().Equal(staleMod) {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		content2, readErr := os.ReadFile(lockPath)
		if readErr != nil || !bytes.Equal(staleContent, content2) {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		if time.Since(info2.ModTime()) <= StaleLockAge {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		_ = os.Remove(lockPath)
		continue
	}
	if f == nil {
		return fmt.Errorf("could not acquire credentials lock at %s; if no other daco process is running, delete that file and retry", lockPath)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(lockPath)
	}()
	return fn()
}

// Save writes credentials. Callers that already hold WithLock should use SaveUnlocked.
func (s FileStore) Save(session Session) error {
	return s.WithLock(func() error {
		return s.SaveUnlocked(session)
	})
}

func (s FileStore) SaveUnlocked(session Session) error {
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
	return s.WithLock(func() error {
		err := os.Remove(s.Path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
}
