# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Locally is a CLI tool designed to help spin up local development environments including infrastructure. It provides configuration files that can be shared and reproduced across machines. The codebase is written in Go and consists of two main entry points:

1. **CLI Tool** (`cmd/cli/main.go`) - Main command-line interface
2. **API Server** (`cmd/api/main.go`) - REST API server with authentication, real-time events, and database support

## Build and Development Commands

### Building the Project
```bash
# Build all packages (verify compilation)
go build -v ./...

# Build CLI executable
go build -o locally ./cmd/cli

# Build API server executable  
go build -o locally-api ./cmd/api
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/environment
```

**Note**: The codebase is currently in a refactoring state with compilation errors. Some tests may fail due to missing AppContext parameters in environment service calls.

### Development
```bash
# Run CLI in debug mode
go run cmd/cli/main.go --debug [command]

# Run API server
go run cmd/api/main.go

# Format code
go fmt ./...

# Run linter (if available)
golangci-lint run
```

## Architecture

### Core Components

#### Application Context (`internal/appctx/`)
- **AppContext**: Enhanced context.Context with request ID, user ID, tenant ID, metadata, and diagnostics
- Thread-safe with structured logging support
- Used throughout the application for tracing and metadata propagation

#### Environment Service (`internal/environment/`)
- Variable interpolation and replacement using `${VAR}` syntax
- Vault integration for secure secret management
- Functions for random value generation
- **Important**: Service methods require `*appctx.AppContext` as first parameter

#### Authentication (`internal/auth/`)
- JWT-based authentication with API key support
- User management with database persistence
- Middleware for API endpoint protection

#### Database (`internal/database/`)
- GORM-based ORM with PostgreSQL and SQLite support
- Data stores for auth, messages, and other entities
- Database migrations and seeding

#### Workers and Lanes (`internal/lanes/`)
- Pipeline execution system with multiple worker types:
  - BashWorker: Execute shell commands
  - CurlWorker: HTTP requests
  - DockerWorker: Container operations
  - GitWorker: Git operations
  - NPMWorker: Node.js package management
  - SqlWorker: Database operations
- **Important**: Worker parameters use environment service for variable replacement

#### Events (`internal/events/`)
- Real-time event system with WebSocket support
- Event hub for system-wide notifications
- API endpoints for event streaming

#### Configuration (`internal/config/`)
- Multi-provider configuration system (file, environment, flags)
- Context-based configuration management
- Support for YAML configuration files

### Key Patterns

1. **Dependency Injection**: Services are initialized and passed to handlers
2. **Context Propagation**: AppContext flows through all operations
3. **Structured Logging**: Logrus with contextual fields
4. **Service Singletons**: Many services use singleton pattern with initialization
5. **Middleware**: Authentication and request processing middleware
6. **Vault System**: Pluggable secret management

### Configuration Structure

The application uses a hierarchical configuration system:
- `configuration/` - Template configurations
- `contexts/` - Environment-specific contexts
- `services/` - Service definitions (backends, infrastructure, etc.)

## Common Operations

### CLI Commands
```bash
# Environment operations
./locally env [subcommand]

# Docker operations  
./locally docker [subcommand]

# Azure Key Vault operations
./locally keyvault [subcommand]

# Infrastructure operations
./locally infrastructure [subcommand] [stack]

# Lanes (pipeline) operations
./locally lanes [subcommand]

# Configuration management
./locally config [subcommand]
```

### API Endpoints
- `/api/auth/*` - Authentication endpoints
- `/api/events/*` - Event streaming
- `/api/environment/*` - Environment variable management
- `/api/workers/*` - Worker/message management

## Development Notes

### Current State
- The codebase is in active refactoring (branch: `refactor-wip-021024`)
- Some compilation errors exist due to AppContext parameter changes
- Tests are partially working

### Testing Strategy
- Uses `testify` for assertions
- Test files follow `*_test.go` convention
- Integration tests for API endpoints
- Unit tests for core services

### Database Support
- PostgreSQL for production
- SQLite for development/testing
- Database configuration via environment variables

### Security
- JWT tokens for API authentication
- API key authentication support
- Encryption service for sensitive data
- Azure Key Vault integration

## Dependencies
- **Web Framework**: Gorilla Mux for HTTP routing
- **Database**: GORM with PostgreSQL/SQLite drivers
- **Authentication**: JWT tokens, bcrypt for passwords
- **Logging**: Logrus with structured logging
- **Docker**: Docker SDK for container operations
- **Azure**: Azure SDK for Key Vault and CLI integration
- **Git**: go-git for Git operations
- **Testing**: testify for test assertions

## File Structure
- `cmd/` - Main entry points (CLI and API)
- `internal/` - Internal application packages
- `pkg/` - Public packages (interfaces, utilities)
- `configuration/` - Configuration templates
- `docs/` - Documentation files