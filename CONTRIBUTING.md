# Contributing to Daco CLI

Thank you for your interest in contributing to Daco CLI! This document covers everything you need to know.

## FAQ

| Question | Answer |
|----------|--------|
| Do you have a Code of Conduct? | Yes, please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| I want to contribute! | Amazing! Head over to [open issues](https://github.com/dacolabs/cli/issues) |
| Can I talk to someone? | Open a [GitHub Discussion](https://github.com/dacolabs/cli/discussions) for questions or ideas |
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

1. **Simplicity Over Cleverness** - Clear, readable code wins over clever abstractions
2. **Idiomatic Patterns** - Follow established conventions of the language and ecosystem
3. **Minimal Dependencies** - Prefer standard library; add dependencies deliberately
4. **Defensive Programming** - Validate inputs, handle errors explicitly, fail gracefully

Daco CLI adheres to the [12-Factor App](https://12factor.net) principles to the extent a CLI tool can.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.23.0+ | Build and run |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | latest | Code linting |
| [GoReleaser](https://goreleaser.com/install/) | v2+ | Local release testing (optional) |
| Make | any | Build automation |

## Development Setup

```bash
# Clone the repository
git clone git@github.com:dacolabs/cli.git
cd cli

# Set up development environment (configures git hooks)
make setup

# Build the project
make build

# Run tests
make test
```

## Project Structure

```
.
├── cmd/daco/              # CLI entry point
│   ├── main.go            # Main function
│   └── internal/          # Application bootstrap
├── internal/              # Private packages
│   ├── commands/          # CLI command definitions
│   ├── config/            # Configuration handling
│   ├── jschema/           # JSON Schema utilities
│   ├── opendpi/           # OpenDPI parsing
│   ├── prompts/           # Interactive prompts
│   ├── session/           # Session management
│   ├── translate/         # Schema translators
│   └── version/           # Version information
├── docs/                  # Documentation
├── .github/               # GitHub configuration
│   └── workflows/         # CI/CD pipelines
├── .githooks/             # Git hooks
├── Makefile               # Build automation
├── .goreleaser.yaml       # Release configuration
└── .golangci.yml          # Linter configuration
```

## Development Workflow

### Building

```bash
make build                    # Build the binary
make run ARGS="init --help"   # Run directly
make install                  # Install to $GOPATH/bin
```

### Testing

```bash
make test   # Run all tests with race detection
```

### Code Quality

```bash
make lint     # Run linter
make format   # Format code
```

### Local Release Testing

```bash
make release-snapshot   # Build release artifacts without publishing
ls -la dist/            # Check generated binaries
```

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
git commit -m "feat: add ports translate command"
git commit -m "feat(commands): add connections list command"
git commit -m "fix: handle empty schema gracefully"
git commit -m "feat!: rename config file to daco.yaml"  # Breaking change
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
git commit -m "feat: add new feature"

# Keep your fork up to date:
git fetch upstream
git rebase upstream/main

# Push and create a pull request:
git push origin my-feature
```

## CI/CD Pipeline

### Continuous Integration

On every push to `main` and on pull requests, CI runs:

1. **Lint** - golangci-lint with project configuration
2. **Test** - `go test -v -race ./...`
3. **Build** - Cross-platform build verification (linux, darwin, windows)

All checks must pass before merging.

### Release Process

Releases are triggered manually via GitHub Actions:

1. Go to **Actions > Release** on GitHub
2. Click **Run workflow**
3. Enter the version (e.g., `0.1.0` or `0.1.0-alpha.1`)
4. Click **Run workflow**

The release workflow:
1. Validates the version format
2. Runs tests
3. Creates a git tag
4. Builds binaries for all platforms via GoReleaser
5. Generates changelog from conventional commits
6. Creates a GitHub release with binaries
7. Updates Homebrew tap and Scoop bucket (stable releases only)

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
| `setup` | Set up development environment (git hooks) |
| `build` | Build the CLI binary |
| `run` | Run the CLI (use `ARGS="..."` for arguments) |
| `install` | Install to `$GOPATH/bin` |
| `test` | Run tests with race detection |
| `lint` | Run golangci-lint |
| `format` | Format code with gofmt and goimports |
| `release-snapshot` | Build release artifacts locally |
| `clean` | Remove build artifacts |

## Communication

We use GitHub for all project communication:

- [GitHub Issues](https://github.com/dacolabs/cli/issues) for bug reports and feature requests
- [GitHub Discussions](https://github.com/dacolabs/cli/discussions) for questions, ideas, and general conversation

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating.
