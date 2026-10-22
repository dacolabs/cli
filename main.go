package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dacolabs/cli/internal/apply"
	"github.com/dacolabs/cli/internal/catalogapi"
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
	case "apply":
		return cmdApply(context.Background(), args[1:])
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
	client, err := catalogapi.NewAuthenticatedClient(base, token, http.DefaultClient)
	if err != nil {
		return err
	}
	pageSize := int32(5)
	res, err := client.ListDatasetsWithResponse(ctx, &catalogapi.ListDatasetsParams{PageSize: &pageSize})
	if err != nil {
		return err
	}
	if res.StatusCode() < 200 || res.StatusCode() >= 300 {
		return fmt.Errorf("catalog %s: %s", res.Status(), strings.TrimSpace(string(res.Body)))
	}
	fmt.Println(string(res.Body))
	return nil
}

func cmdApply(ctx context.Context, args []string) error {
	var files []string
	dryRun := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "-f", "--filename":
			i++
			if i >= len(args) {
				return fmt.Errorf("apply: %s requires a path", args[i-1])
			}
			files = append(files, args[i])
		case "--help", "-h":
			fmt.Print(applyUsage)
			return nil
		default:
			return fmt.Errorf("apply: unknown argument %q\n%s", args[i], applyUsage)
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("apply: at least one -f path is required\n%s", applyUsage)
	}

	units, err := apply.Load(files)
	if err != nil {
		return err
	}

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
	client, err := catalogapi.NewAuthenticatedClient(base, token, http.DefaultClient)
	if err != nil {
		return err
	}

	results, err := apply.Run(ctx, client, units, dryRun)
	if err != nil {
		return err
	}
	var created, patched, unchanged, failed int
	for _, r := range results {
		line := fmt.Sprintf("%s  %s@%s", r.Action, r.Urn, r.Version)
		if r.Message != "" {
			line += "  " + r.Message
		}
		fmt.Println(line)
		switch r.Action {
		case apply.ActionCreated:
			created++
		case apply.ActionPatched:
			patched++
		case apply.ActionUnchanged:
			unchanged++
		default:
			failed++
		}
	}
	fmt.Fprintf(os.Stderr, "summary: created=%d patched=%d unchanged=%d error=%d\n", created, patched, unchanged, failed)
	if failed > 0 {
		return fmt.Errorf("apply finished with %d error(s)", failed)
	}
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

// Set at link time by GoReleaser: -X main.version={{.Version}}
var version = "0.0.0-dev"

const usage = `daco - Daco Catalog CLI

Usage:
  daco login
  daco logout
  daco whoami
  daco datasets
  daco apply -f <file|dir> [-f ...] [--dry-run]
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

const applyUsage = `Apply Dataset YAML (create/update; no prune):
  daco apply -f datasets.yaml
  daco apply -f ./manifests --dry-run

Documents use kind: Dataset with Catalog fields (urn, version, title, description, metadata, contract).
Multiple documents may be separated with ---. Conflicting urn@version declarations fail before API calls.
`
