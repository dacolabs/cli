package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dacolabs/cli/internal/catalog"
	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/cli/internal/creds"
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/cli/internal/workosauth"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return fmt.Errorf("missing command")
	}
	switch args[0] {
	case "--help", "-h", "help":
		fmt.Print(usage)
		return nil
	case "--version", "-v", "version":
		fmt.Printf("daco %s\n", version)
		return nil
	case "login":
		return cmdLogin(context.Background())
	case "logout":
		return cmdLogout()
	case "whoami":
		return cmdWhoami(context.Background())
	case "datasets":
		return cmdDatasets(context.Background())
	default:
		fmt.Print(usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func credentialStore() (creds.FileStore, error) {
	path, err := creds.DefaultPath()
	if err != nil {
		return creds.FileStore{}, err
	}
	return creds.FileStore{Path: path}, nil
}

func cmdLogin(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.RequireClientID(); err != nil {
		return err
	}
	store, err := credentialStore()
	if err != nil {
		return err
	}
	client := workosauth.New(cfg.AuthAPI, cfg.ClientID, http.DefaultClient)
	auth, err := client.AuthorizeDevice(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Open %s and confirm code %s\n", auth.VerificationURIComplete, auth.UserCode)
	if auth.VerificationURIComplete == "" {
		fmt.Fprintf(os.Stderr, "Open %s and enter code %s\n", auth.VerificationURI, auth.UserCode)
	}
	tok, err := client.PollDevice(ctx, auth)
	if err != nil {
		return err
	}
	if tok.OrganizationID == "" {
		return fmt.Errorf("login succeeded without an organization; select an organization during device confirmation and try again")
	}
	session := creds.Session{
		ClientID:       cfg.ClientID,
		AccessToken:    tok.AccessToken,
		RefreshToken:   tok.RefreshToken,
		OrganizationID: tok.OrganizationID,
		UserID:         tok.UserID,
		UserEmail:      tok.UserEmail,
		ExpiresAt:      tok.ExpiresAt,
		AuthAPI:        cfg.AuthAPI,
		BaseURL:        cfg.BaseURL,
	}
	if err := store.Save(session); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Logged in as %s (org %s). Credentials expire around %s.\n",
		displayIdentity(tok.UserEmail, tok.UserID), tok.OrganizationID, tok.ExpiresAt.Local().Format(time.RFC822))
	return nil
}

func cmdLogout() error {
	store, err := credentialStore()
	if err != nil {
		return err
	}
	if err := store.Clear(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Logged out.")
	return nil
}

func cmdWhoami(ctx context.Context) error {
	store, err := credentialStore()
	if err != nil {
		return err
	}
	_, sess, err := session.AccessToken(ctx, store, http.DefaultClient)
	if err != nil {
		return err
	}
	fmt.Printf("user:  %s\n", displayIdentity(sess.UserEmail, sess.UserID))
	fmt.Printf("org:   %s\n", sess.OrganizationID)
	fmt.Printf("exp:   %s\n", sess.ExpiresAt.Local().Format(time.RFC3339))
	return nil
}

func cmdDatasets(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	store, err := credentialStore()
	if err != nil {
		return err
	}
	token, sess, err := session.AccessToken(ctx, store, http.DefaultClient)
	if err != nil {
		return err
	}
	base := cfg.BaseURL
	if base == "" {
		base = sess.BaseURL
	}
	if base == "" {
		return fmt.Errorf("set DACO_BASE_URL to your Catalog API origin")
	}
	body, _, err := catalog.Client{BaseURL: base, Token: token}.Get(ctx, "/api/catalog/datasets?pageSize=5")
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func displayIdentity(email, id string) string {
	email = strings.TrimSpace(email)
	if email != "" {
		return email
	}
	if id != "" {
		return id
	}
	return "(unknown)"
}

const version = "0.0.0-dev"

const usage = `daco - Daco Catalog CLI

Usage:
  daco login
  daco logout
  daco whoami
  daco datasets
  daco --help
  daco --version

Authenticate with WorkOS AuthKit device authorization (human session), then call the Catalog API.

Environment:
  DACO_ENV         staging (default) or production
  DACO_BASE_URL    Catalog API origin (defaults with DACO_ENV)
  DACO_CLIENT_ID   Override AuthKit client ID (public; baked in for staging)
  DACO_AUTH_API    AuthKit API host (default https://api.workos.com)
  DACO_CREDENTIALS_FILE  Override credentials path
`
