# Daco CLI

Swiss army knife for data product projects.

## Installation

```bash
go install github.com/dacolabs/cli/cmd/daco@latest
```

Or build from source:

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

## Development

```bash
# Run tests
make test

# Lint code
make lint

# Format code
make format
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.
