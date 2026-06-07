# Contributing to Daco

Thank you for your interest in contributing to Daco! This document covers everything you need to know to work on the **`daco` CLI**.

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
5. **Thin drivers, shared engine** — Business logic lives in `internal/cli/engine`; the cobra commands and the TUI are pure drivers over the same operations

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

# Build the CLI
make build

# Run tests
make test
```

After `make build`, the binary lands in `bin/`:

- `bin/daco` — the CLI

## Project Structure

```
.
├── cmd/
│   └── daco/             # CLI entrypoint
│       ├── app/          # Bootstrap (registers commands + TUI runner)
│       └── main.go
├── internal/
│   ├── cli/              # CLI drivers and shared state
│   │   ├── commands/     # cobra commands (one file per noun)
│   │   ├── engine/       # Business logic (Input/Valid/Output pattern)
│   │   ├── settings/     # User (~/.daco/settings.yaml) and project (daco.yaml)
│   │   └── tui/          # rivo/tview interactive driver
│   ├── opendpi/          # OpenDPI document/schema parsing with $ref preservation
│   ├── translate/        # Schema translation pipeline
│   │   ├── registry/     # Default() map of format name → translator
│   │   └── <target>/     # One package per format (avro, pyspark, gotypes, …)
│   └── version/          # Build-time version information (ldflag-injected)
├── opendpi/              # Vendored OpenDPI specification
├── docs/                 # User-facing documentation
├── landing/              # Marketing site (Next.js — not part of the Go build)
├── .github/workflows/    # CI and release pipelines
├── Makefile              # Build automation
├── .goreleaser.yaml      # Release configuration
└── .golangci.yml         # Linter configuration
```

See [CLAUDE.md](CLAUDE.md) for a deeper architectural tour.

## Development Workflow

### Building

```bash
make build                          # Build the CLI (alias for build-cli)
make build-cli                      # Build the CLI explicitly

make run-cli ARGS="init --help"     # Run the CLI without building

make install                        # Install to $GOPATH/bin (alias for install-cli)
make install-cli                    # Install the CLI explicitly
```

### Testing

```bash
make test            # Run all tests with race detection + per-package coverage
make test-coverage   # Run tests and emit an HTML coverage report (coverage.html)
```

Run a single test or package:

```bash
go test -race -run TestPortsTranslate ./internal/cli/engine
go test -race -run TestTranslate       ./internal/translate/pyspark
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

Each archive in `dist/` contains the `daco` binary for that platform.

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages. The release pipeline uses commit types to generate the changelog automatically.

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
- [ ] The binary still builds (`make build`)
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
3. **Build** — Cross-platform build verification (linux, darwin, windows × amd64, arm64; windows/arm64 excluded)

All checks must pass before merging.

### Release Process

Releases are triggered manually via GitHub Actions.

1. Go to **Actions → Release** on GitHub
2. Click **Run workflow**
3. Enter the version (e.g., `0.1.0` or `0.1.0-alpha.1`)
4. Click **Run workflow**

The release workflow:

1. Validates the version format (semver, with optional `-alpha.N` / `-beta.N` / `-rc.N` suffix)
2. Runs the test suite
3. Creates and pushes a git tag
4. Builds `daco` for every supported platform via GoReleaser
5. Generates a changelog from conventional commits
6. Creates a GitHub release with archives
7. Opens PRs against the Homebrew tap and Scoop bucket (stable releases only — prereleases are skipped automatically)

### Versioning Strategy

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
| `build` | Build the CLI (alias for `build-cli`) |
| `build-cli` | Build the CLI binary |
| `run-cli` | Run the CLI (use `ARGS="..."` for arguments) |
| `install` | Install to `$GOPATH/bin` (alias for `install-cli`) |
| `install-cli` | Install the CLI to `$GOPATH/bin` |
| `test` | Run tests with race detection and coverage |
| `test-coverage` | Run tests and emit an HTML coverage report |
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
