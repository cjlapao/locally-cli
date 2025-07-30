# Locally

Locally is a command line tool designed to help spin up a local environment including the infrastructure. The concept is simple: have configuration files created by each team that can be shareable and reproducible from machine to machine and deploy the bare minimum infrastructure.

## Quick Start

The project uses a comprehensive Makefile for easy cross-platform development. Get started quickly:

```bash
# Initialize the project (creates environment files and downloads dependencies)
make init

# Build the API (default)
make build

# Build the CLI
make TARGET=cli build

# Run tests
make test

# Show all available commands
make help
```

## Prerequisites

### Required Tools

**Make**: The project uses a cross-platform Makefile for all operations.

- **Linux/macOS**: Usually pre-installed. If not: `sudo apt-get install make` (Ubuntu/Debian) or `brew install make` (macOS)
- **Windows**: Install via [Chocolatey](https://chocolatey.org/install) (`choco install make`) or [Scoop](https://scoop.sh/) (`scoop install make`)

**Go**: Download from [go.dev/dl](https://go.dev/dl/) and follow the installation guide for your platform.

**Docker** (for API development): Download from [docker.com](https://www.docker.com/products/docker-desktop/).

### VS Code Setup

The project is configured for VS Code development:

1. **Install VS Code**: Download from [code.visualstudio.com](https://code.visualstudio.com/)
2. **Install Go Extension**: Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) by Google
3. **Install Recommended Extensions**: VS Code will prompt you to install additional Go-related extensions
4. **Environment Files**: The project uses `.api.env` and `.cli.env` files that VS Code automatically recognizes

**VS Code Go Setup Guide**: [code.visualstudio.com/docs/languages/go](https://code.visualstudio.com/docs/languages/go)

## Environment Configuration

The project uses environment-specific configuration files that VS Code automatically recognizes:

### API Environment (`.api.env`)

Located at `cmd/api/.api.env`, this file contains API-specific configuration:

```bash
# Copy the template to create your environment file
cp cmd/api/env.template cmd/api/.api.env

# Edit the file with your configuration
# VS Code will automatically recognize this file
```

### CLI Environment (`.cli.env`)

Located at `cmd/cli/.cli.env`, this file contains CLI-specific configuration:

```bash
# Create the CLI environment file
touch cmd/cli/.cli.env

# Edit the file with your configuration
# VS Code will automatically recognize this file
```

## Development Workflow

### Project Initialization

```bash
# Initialize for API (default)
make init

# Initialize for CLI
make TARGET=cli init

# Complete setup (init + dependencies)
make setup
```

### Building

The project supports both API and CLI targets:

```bash
# Build API (default)
make build

# Build CLI
make TARGET=cli build

# Build for specific platform
make build-linux
make build-macos
make build-windows

# Build for all platforms
make build-all
```

### Testing

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make test-coverage
```

### Development Mode

```bash
# Run API in development mode
make dev

# Build and run with Docker (API only)
make docker-full
```

### Code Quality

```bash
# Format code
make fmt

# Run linting
make lint

# Full build with linting and testing
make full-build
```

## Docker Development (API Only)

The API includes Docker support for containerized development:

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Stop container
make docker-stop

# View logs
make docker-logs

# Clean Docker artifacts
make docker-clean
```

### Docker Registry

Push to the project's Docker registry:

```bash
# Login to registry
make docker-login

# Build and push (both latest and version tags)
make docker-push

# Push beta version (timestamp format)
make docker-push-beta
```

## Advanced Usage

### Target Selection

The Makefile supports multiple targets with the `TARGET` parameter:

```bash
# API (default)
make build
make TARGET=api build

# CLI
make TARGET=cli build
make TARGET=cli test
```

### Cross-Platform Building

```bash
# Build for specific platform
GOOS=linux GOARCH=amd64 make build-cross
GOOS=darwin GOARCH=arm64 make build-cross
GOOS=windows GOARCH=amd64 make build-cross

# Build for all platforms
make build-all
```

### Custom Build Parameters

```bash
# Build with specific parameters
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 make build-cross

# Build without CGO (not recommended - SQLite won't work)
CGO_ENABLED=0 make build-cross
```

## Troubleshooting

### Common Issues

**Make not found on Windows**: Install via Chocolatey (`choco install make`) or Scoop (`scoop install make`)

**CGO compilation errors**: Ensure you have the appropriate C compiler for your target platform

**Docker commands fail for CLI**: Docker targets are only available for API (`TARGET=api`)

**Environment files not recognized**: Ensure files are named `.api.env` and `.cli.env` in their respective directories

### Getting Help

- **Troubleshooting Guide**: [docs/troubleshooting.md](./docs/troubleshooting.md)
- **Makefile Help**: `make help`
- **Team Channel**: [Microsoft Teams Locally Channel](https://teams.microsoft.com/l/channel/19%3a98b5d070649f442ab23b247ec5858e16%40thread.skype/locally%2520-%2520Also%2520called%2520Locally?groupId=cd5ee759-4aef-4928-95f4-b8c658c5d0db&tenantId=e5208e76-dd12-47f0-9541-c9b45afaffe6)

## Project Structure

```
locally-cli/
├── cmd/
│   ├── api/          # API service
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── .api.env  # API environment (VS Code recognized)
│   │   └── env.template
│   └── cli/          # CLI service
│       ├── main.go
│       └── .cli.env  # CLI environment (VS Code recognized)
├── internal/         # Internal packages
├── scripts/          # Cross-platform scripts
├── docs/             # Documentation
├── Makefile          # Cross-platform build system
└── README.md         # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Ensure code quality: `make lint`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
