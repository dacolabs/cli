# Daco CLI

Swiss army knife for data product projects.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install dacolabs/tap/daco
```

### Scoop (Windows)

```powershell
scoop bucket add dacolabs https://github.com/dacolabs/scoop-bucket.git
scoop install daco
```

### Go Install

```bash
go install github.com/dacolabs/cli/cmd/daco@latest
```

### Manual Download

Download the appropriate binary from the [releases page](https://github.com/dacolabs/cli/releases).

### Build from Source

```bash
git clone https://github.com/dacolabs/cli.git
cd cli
make build
./bin/daco --help
```

## Usage

```bash
# Initialize a new data product
daco init

# Show help
daco --help
```

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, project structure, and guidelines.

## License

This project is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.
