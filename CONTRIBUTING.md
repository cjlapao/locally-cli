# Contributing to Locally CLI

[fork]: https://github.com/cjlapao/locally-cli/fork
[pr]: https://github.com/cjlapao/locally-cli/compare
[code-of-conduct]: CODE_OF_CONDUCT.md

Hi there! We're thrilled that you'd like to contribute to Locally CLI. Your help
is essential for keeping it great and making local development easier for everyone.

We accept pull requests for bug fixes and features where we've discussed the
approach in an issue and given the go-ahead for a community member to work on
it. We'd also love to hear about ideas for new features as issues.

## How to Contribute

### Before You Start

Please do:

* Check existing issues to verify that the [bug][bug issues] or
  [feature request][feature request issues] has not already been submitted.
* Open an issue if things aren't working as expected.
* Open an issue to propose a significant change.
* Open a pull request to fix a bug.
* Open a pull request to fix documentation about a command.

Please avoid:

* Opening pull requests for issues marked `needs-design`, `needs-investigation`,
  or `blocked`.
* Making changes without discussing them first in an issue (for significant features).

Contributions to this project are released to the public under the
[project's open source license](LICENSE).

Please note that this project is released with a
[Contributor Code of Conduct][code-of-conduct]. By participating in this project
you agree to abide by its terms.

## Prerequisites for Development

These are one-time installations required to be able to test your changes locally
as part of the pull request (PR) submission process.

### Required Tools

1. **Go**: Install from [go.dev/dl](https://go.dev/dl/)
2. **Make**: 
   - **Linux/macOS**: Usually pre-installed. If not: `sudo apt-get install make` (Ubuntu/Debian) or `brew install make` (macOS)
   - **Windows**: Install via [Chocolatey](https://chocolatey.org/install) (`choco install make`) or [Scoop](https://scoop.sh/) (`scoop install make`)
3. **VS Code** (recommended): Download from [code.visualstudio.com](https://code.visualstudio.com/)
4. **Docker** (for API development): Download from [docker.com](https://www.docker.com/products/docker-desktop/)

### Project Setup

1. **Fork and clone** the repository
2. **Initialize the project**:
   ```bash
   # For API development (default)
   make init
   
   # For CLI development
   make TARGET=cli init
   ```
3. **Install dependencies**:
   ```bash
   make deps
   ```

## Development Workflow

### Environment Configuration

The project uses environment-specific configuration files:

- **API**: `cmd/api/.api.env` (created from `cmd/api/env.template`)
- **CLI**: `cmd/cli/.cli.env` (created during init)

VS Code will automatically recognize these files for IntelliSense and debugging.

### Building and Testing

#### Quick Development

```bash
# Build the current target (API or CLI)
make build

# Run tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make test-coverage
```

#### Target-Specific Development

```bash
# API development
make TARGET=api build
make TARGET=api test
make TARGET=api dev

# CLI development
make TARGET=cli build
make TARGET=cli test
```

#### Code Quality

```bash
# Format code
make fmt

# Run linting
make lint

# Full build with linting and testing
make full-build
```

### Docker Development (API Only)

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# View logs
make docker-logs

# Stop container
make docker-stop
```

### Cross-Platform Building

```bash
# Build for specific platform
make build-linux
make build-macos
make build-windows

# Build for all platforms
make build-all

# Custom platform build
GOOS=linux GOARCH=amd64 make build-cross
```

## Submitting a Pull Request

1. **Fork and clone** the repository
2. **Set up your development environment**:
   ```bash
   make init
   make deps
   ```
3. **Create a new branch**: `git checkout -b my-feature-name`
4. **Make your changes** and ensure they work for both API and CLI targets
5. **Add tests** for new functionality
6. **Run quality checks**:
   ```bash
   make test
   make lint
   make full-build
   ```
7. **Push to your fork** and [submit a pull request][pr]

### Pull Request Guidelines

Here are a few things you can do that will increase the likelihood of your pull
request being accepted:

* **Test both targets**: Ensure your changes work for both API (`TARGET=api`) and CLI (`TARGET=cli`)
* **Write tests**: Add tests for new functionality
* **Keep changes focused**: If there are multiple changes that are not dependent upon each other, consider submitting them as separate pull requests
* **Write good commit messages**: Follow the [conventional commit format](https://www.conventionalcommits.org/)
* **Update documentation**: If you're adding new features, update the README and relevant documentation
* **Check environment files**: Ensure any new environment variables are documented in the appropriate `.env` template

### Commit Message Format

We use conventional commit messages:

```
type(scope): description

[optional body]

[optional footer]
```

Examples:
- `feat(api): add new authentication endpoint`
- `fix(cli): resolve build error on Windows`
- `docs: update README with new installation steps`
- `test: add unit tests for validation package`

## Project Structure

```
locally-cli/
├── cmd/
│   ├── api/          # API service
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── .api.env  # API environment
│   │   └── env.template
│   └── cli/          # CLI service
│       ├── main.go
│       └── .cli.env  # CLI environment
├── internal/         # Internal packages
├── scripts/          # Cross-platform scripts
├── docs/             # Documentation
├── Makefile          # Cross-platform build system
└── README.md         # Main documentation
```

## Resources

* [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
* [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
* [GitHub Help](https://help.github.com)
* [Conventional Commits](https://www.conventionalcommits.org/)
* [VS Code Go Setup](https://code.visualstudio.com/docs/languages/go)

## Getting Help

* **Issues**: [GitHub Issues](https://github.com/cjlapao/locally-cli/issues)
* **Discussions**: [GitHub Discussions](https://github.com/cjlapao/locally-cli/discussions)
* **Feedback**: [feedback@locally.cloud](mailto:feedback@locally.cloud)
* **Licensing**: [licensing@locally.cloud](mailto:licensing@locally.cloud)

[bug issues]: https://github.com/cjlapao/locally-cli/labels/bug
[feature request issues]: https://github.com/cjlapao/locally-cli/labels/feature-request
