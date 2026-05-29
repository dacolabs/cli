# Contributing to Daco

Thank you for your interest in contributing to Daco! This document covers everything you need to know to work on the **`daco` CLI** and the **`daco-api` server**, which live in the same repository.

## FAQ

| Question | Answer |
|----------|--------|
| Do you have a Code of Conduct? | Yes, please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| I want to contribute! | Amazing! Head over to [open issues](https://github.com/dacolabs/cli/issues) |
| Can I talk to someone? | Open a [GitHub Discussion](https://github.com/dacolabs/cli/discussions) for questions or ideas |
| How are vulnerabilities reported? | Privately — see [SECURITY.md](SECURITY.md) |
| What's the license? | [Apache 2.0](LICENSE) |

## How Can I Contribute?

- **Star the repository** and help spread the word
- **Answer questions** in [GitHub Discussions](https://github.com/dacolabs/cli/discussions) to help other users
- **Report bugs** with detailed reproduction steps
- **Fix bugs** by submitting pull requests
- **Improve documentation** by fixing typos, clarifying explanations, or adding examples
- **Add new features** after discussing with maintainers
- **Write tests** to improve code coverage

## Design Principles

1. **Simplicity Over Cleverness** — Clear, readable code wins over clever abstractions
2. **Idiomatic Patterns** — Follow established conventions of Go and the surrounding ecosystem
3. **Minimal Dependencies** — Prefer the standard library; add dependencies deliberately
4. **Defensive Programming** — Validate inputs, handle errors explicitly, fail gracefully
5. **Shared core, thin entrypoints** — Business logic lives in `internal/`; the binaries in `cmd/` are wiring

The CLI and the API are different surfaces over the same domain model. When you add functionality, prefer putting the logic in a shared `internal/` package and exposing it from both surfaces where it makes sense.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.26+ | Build and run |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | v2.1+ | Code linting |
| [GoReleaser](https://goreleaser.com/install/) | v2+ | Local release testing (optional) |
| Make | any | Build automation |

## Development Setup

```bash
# Clone the repository
git clone git@github.com:dacolabs/cli.git
cd cli

# Set up development environment (configures git hooks)
make setup

# Build both binaries
make build

# Run tests
make test
```

After `make build`, the binaries land in `bin/`:

- `bin/daco` — the CLI
- `bin/daco-api` — the HTTP API server

## Project Structure

```
.
├── cmd/
│   ├── cli/              # `daco` CLI entrypoint
│   │   ├── app/          # Bootstrap (command registration, run loop)
│   │   └── main.go
│   └── api/              # `daco-api` HTTP server entrypoint
│       ├── app/          # Server bootstrap (config, lifecycle)
│       └── main.go
├── internal/
│   ├── api/              # HTTP handlers and middleware (used by daco-api)
│   ├── cli/              # CLI commands, config, prompts, session, TUI
│   ├── env/              # Environment variable parsing helpers
│   ├── jschema/          # JSON Schema utilities
│   ├── opendpi/          # OpenDPI parsing
│   ├── services/         # Shared application services
│   ├── storage/          # Data access
│   ├── telemetry/        # OpenTelemetry + slog setup
│   ├── translate/        # Schema translators (avro, pyspark, scala, …)
│   └── version/          # Build-time version information
├── db/
│   ├── migrations/       # SQL migrations
│   └── queries/          # sqlc query definitions
├── docs/                 # User-facing documentation
├── landing/              # Marketing site (Next.js — not part of the Go build)
├── opendpi/              # OpenDPI assets / fixtures
├── .github/workflows/    # CI and release pipelines
├── .githooks/            # Git hooks installed by `make setup`
├── Makefile              # Build automation
├── .goreleaser.yaml      # Release configuration
└── .golangci.yml         # Linter configuration
```

## Development Workflow

### Building

```bash
make build                          # Build both binaries
make build-cli                      # CLI only
make build-api                      # API server only

make run-cli ARGS="init --help"     # Run the CLI without building
make run-api                        # Run the API server (defaults: port 8080)

make install                        # Install both to $GOPATH/bin
```

### Testing

```bash
make test     # Run all tests with race detection
```

### Code Quality

```bash
make lint     # Run golangci-lint
make format   # gofmt -s + goimports
```

### Local Release Testing

```bash
make release-snapshot   # Build release artifacts without publishing
ls -la dist/            # Inspect generated binaries and archives
```

Each archive in `dist/` contains **both** `daco` and `daco-api` for that platform.

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages. A pre-commit hook validates this automatically after running `make setup`.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `chore` | Maintenance tasks |
| `test` | Adding or updating tests |
| `refactor` | Code refactoring |
| `ci` | CI/CD changes |
| `build` | Build system changes |
| `perf` | Performance improvements |
| `style` | Code style changes (formatting) |

### Examples

```bash
git commit -m "feat(cli): add ports translate command"
git commit -m "feat(api): add /health/ready endpoint"
git commit -m "fix(translate/pydantic): handle empty schema gracefully"
git commit -m "feat!: rename config file to daco.yaml"   # Breaking change
```

## Contributing Code Changes

### Process

1. **Open an issue** to discuss the bug or feature
2. **Get approval** from a maintainer before starting significant work
3. **Submit a pull request** once the work is ready for review

### Pull Request Checklist

Before submitting a pull request, ensure:

- [ ] An issue was opened and discussed (for non-trivial changes)
- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Both binaries still build (`make build`)
- [ ] Documentation is updated if applicable
- [ ] Commit messages follow Conventional Commits
- [ ] The PR description clearly summarizes the changes

### Working with Forks

```bash
# Fork the repository on GitHub, then clone your fork:
git clone git@github.com:<your-username>/cli.git
cd cli

# Add the official repository as upstream:
git remote add upstream git@github.com:dacolabs/cli.git

# Create a feature branch:
git checkout -b my-feature

# Make your changes, then commit:
git add .
git commit -m "feat(cli): add new feature"

# Keep your fork up to date:
git fetch upstream
git rebase upstream/main

# Push and create a pull request:
git push origin my-feature
```

## CI/CD Pipeline

### Continuous Integration

On every push to `main` and on pull requests targeting `main`, CI runs:

1. **Lint** — `golangci-lint` with the project configuration
2. **Test** — `go test -v -race ./...`
3. **Build** — Cross-platform build verification for **both** `daco` and `daco-api` (linux, darwin, windows × amd64, arm64; windows/arm64 excluded)

All checks must pass before merging.

### Release Process

Releases are triggered manually via GitHub Actions and publish `daco` and `daco-api` together under a single version tag.

1. Go to **Actions → Release** on GitHub
2. Click **Run workflow**
3. Enter the version (e.g., `0.1.0` or `0.1.0-alpha.1`)
4. Click **Run workflow**

The release workflow:

1. Validates the version format (semver, with optional `-alpha.N` / `-beta.N` / `-rc.N` suffix)
2. Runs the test suite
3. Creates and pushes a git tag
4. Builds **both** binaries for every supported platform via GoReleaser
5. Generates a changelog from conventional commits
6. Creates a GitHub release with archives — each archive bundles `daco` and `daco-api`
7. Opens PRs against the Homebrew tap and Scoop bucket (stable releases only)

### Versioning Strategy

`daco` and `daco-api` share a single version. One tag → one release → both binaries at the same version.

| Release Type | Version Format | Example | Published to Homebrew/Scoop |
|--------------|----------------|---------|------------------------------|
| Alpha | `X.Y.Z-alpha.N` | `0.1.0-alpha.1` | No |
| Beta | `X.Y.Z-beta.N` | `0.1.0-beta.1` | No |
| Release Candidate | `X.Y.Z-rc.N` | `0.1.0-rc.1` | No |
| Stable | `X.Y.Z` | `0.1.0` | Yes |

Pre-releases are automatically detected from the version string and marked as such on GitHub.

## Makefile Targets

Run `make help` to see all available targets:

| Target | Description |
|--------|-------------|
| `setup` | Set up development environment (git hooks) |
| `build` | Build both binaries |
| `build-cli` | Build the CLI only |
| `build-api` | Build the API server only |
| `run-cli` | Run the CLI (use `ARGS="..."` for arguments) |
| `run-api` | Run the API server |
| `install` | Install both binaries to `$GOPATH/bin` |
| `install-cli` | Install the CLI only |
| `install-api` | Install the API server only |
| `test` | Run tests with race detection |
| `lint` | Run golangci-lint |
| `format` | Format code (gofmt + goimports) |
| `release-snapshot` | Build release artifacts locally (no publish) |
| `clean` | Remove build artifacts |

## Communication

We use GitHub for all project communication:

- [GitHub Issues](https://github.com/dacolabs/cli/issues) for bug reports and feature requests
- [GitHub Discussions](https://github.com/dacolabs/cli/discussions) for questions, ideas, and general conversation

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating.
