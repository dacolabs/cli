# Contributing to Daco CLI

Thank you for your interest in contributing to Daco CLI! This document explains how to contribute to the project.

## FAQ

| Question | Answer |
|----------|--------|
| Do you have a Code of Conduct? | Yes, please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| I want to contribute! | Amazing! Head over to [open issues](https://github.com/dacolabs/cli/issues) |
| Can I talk to someone? | Open a [GitHub Discussion](https://github.com/dacolabs/cli/discussions) for questions or ideas |
| What's the license? | [Apache 2.0](LICENSE) |

## How Can I Contribute?

There are many ways to contribute to Daco CLI:

- **Star the repository** and help spread the word
- **Answer questions** in [GitHub Discussions](https://github.com/dacolabs/cli/discussions) to help other users
- **Report bugs** with detailed reproduction steps
- **Fix bugs** by submitting pull requests
- **Improve documentation** by fixing typos, clarifying explanations, or adding examples
- **Add new features** after discussing with maintainers
- **Write tests** to improve code coverage

## Communication

We use GitHub for all project communication:

- [GitHub Issues](https://github.com/dacolabs/cli/issues) for bug reports and feature requests
- [GitHub Discussions](https://github.com/dacolabs/cli/discussions) for questions, ideas, and general conversation

## Development Setup

```bash
# Clone the repository
git clone git@github.com:dacolabs/cli.git
cd cli

# Install dependencies
go mod download

# Run tests
make test

# Build the project
make build

# Run linter
make lint
```

## Contributing Code Changes

Changes to the codebase should follow this process:

### Process

1. **Open an issue** to discuss the bug or feature
2. **Get approval** from a maintainer before starting significant work
3. **Submit a pull request** once the work is ready for review

## Pull Request Guidelines

Before submitting a pull request, ensure:

- [ ] **An issue was opened** and discussed (for non-trivial changes)
- [ ] **Tests pass** (`make test`)
- [ ] **Linter passes** (`make lint`)
- [ ] **Documentation** is updated if applicable
- [ ] **Commit messages** follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standard
- [ ] **The PR description clearly summarizes** the technical changes and implementation details

## Working with Forks

```bash
# Fork the repository on GitHub, then clone your fork locally:
git clone git@github.com:<your-username>/cli.git
cd cli

# Add the official repository as a remote named 'upstream':
git remote add upstream git@github.com:dacolabs/cli.git

# Create a feature branch:
git checkout -b my-feature

# Make your changes, then:
git add .
git commit -m "feat: <short feature description>"

# Keep your fork up to date:
git fetch upstream
git rebase upstream/main

# Push and create a pull request:
git push origin my-feature
```

## Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating.
