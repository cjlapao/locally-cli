# Makefile for locally-cli

# Go version
GO_VERSION=1.23

# Version management
VERSION_FILE=VERSION
CHANGELOG_FILE=CHANGELOG.md
RELEASE_NOTES_FILE=release_notes.md
REPO_NAME=https://github.com/cjlapao/locally-cli

# Variables
BINARY_NAME=locally
GO=go
GOFMT=gofmt
GOLINT=golangci-lint
GOTEST=$(GO) test
GOVET=$(GO) vet
GOCOVER=$(GO) tool cover
GOCOV=gocov
GOCOVXML=gocov-xml
GOCOVHTML=gocov-html

# Required tools
REQUIRED_TOOLS=swag gocov gocov-xml gosec

# Build targets for different platforms
GOOS_LINUX=linux
GOOS_WINDOWS=windows
GOOS_DARWIN=darwin
GOARCH_AMD64=amd64
GOARCH_ARM64=arm64

# Source directory
SRC_DIR=src

# Output directories
BIN_DIR=bin
DIST_DIR=dist
COVERAGE_DIR=coverage

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Check Go version
.PHONY: check-version
check-version:
	@echo "Checking Go version..."
	@go version | grep -q "go$(GO_VERSION)" || (echo "Error: Required Go version is $(GO_VERSION)" && exit 1)
	@echo "Go version check passed"

# Check and install required tools
.PHONY: check-tools
check-tools:
	@echo "Checking required tools..."
	@for tool in $(REQUIRED_TOOLS); do \
		if ! command -v $$tool >/dev/null 2>&1; then \
			echo "$$tool not found, will be installed"; \
		else \
			echo "$$tool is already installed"; \
		fi; \
	done

.PHONY: install-tools
install-tools: check-tools
	@echo "Installing required Go tools..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@if ! command -v gocov >/dev/null 2>&1; then \
		echo "Installing gocov..."; \
		go install github.com/axw/gocov/gocov@latest; \
	fi
	@if ! command -v gocov-xml >/dev/null 2>&1; then \
		echo "Installing gocov-xml..."; \
		go install github.com/AlekSi/gocov-xml@latest; \
	fi
	@if ! command -v gocov-html >/dev/null 2>&1; then \
		echo "Installing gocov-html..."; \
		go install github.com/matm/gocov-html/cmd/gocov-html@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "All required tools are installed"

# Default target
.PHONY: all
all: check-version check-tools clean lint test build docs

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@mkdir -p $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@cd $(SRC_DIR) && $(GO) build -v -o ../$(BIN_DIR)/$(BINARY_NAME) -ldflags "-X main.version=$(VERSION)"
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@cd $(SRC_DIR) && $(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage: check-tools
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@cd $(SRC_DIR) && $(GOTEST) -coverprofile=../$(COVERAGE_DIR)/coverage.txt -covermode=count -v ./...
	@cd $(COVERAGE_DIR) && $(GOCOVER) -func=coverage.txt
	@echo "Generating HTML coverage report..."
	@cd $(COVERAGE_DIR) && $(GOCOVER) -html=coverage.txt -o coverage.html
	@echo "Generating XML coverage report..."
	@cd $(COVERAGE_DIR) && $(GOCOV) convert coverage.txt | $(GOCOVXML) > coverage.xml
	@echo "Coverage reports generated in $(COVERAGE_DIR): coverage.txt, coverage.html, coverage.xml"

# Open coverage report in browser
.PHONY: coverage-report
coverage-report: test-coverage
	@echo "Opening coverage report in browser..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open $(COVERAGE_DIR)/coverage.html; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		xdg-open $(COVERAGE_DIR)/coverage.html 2>/dev/null || echo "Please open $(COVERAGE_DIR)/coverage.html in your browser"; \
	else \
		start $(COVERAGE_DIR)/coverage.html 2>/dev/null || echo "Please open $(COVERAGE_DIR)/coverage.html in your browser"; \
	fi

# Run linting
.PHONY: lint
lint:
	@echo "Running linters..."
	@cd $(SRC_DIR) && $(GOVET) ./...
	@echo "Checking formatting..."
	@cd $(SRC_DIR) && test -z "$$($(GOFMT) -l .)"

# Install golangci-lint if not present
.PHONY: install-lint
install-lint:
	@which $(GOLINT) >/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

# Run golangci-lint if installed
.PHONY: golangci-lint
golangci-lint: install-lint
	@echo "Running golangci-lint..."
	@cd $(SRC_DIR) && $(GOLINT) run ./...

# Cross-compile for different platforms
.PHONY: dist
dist: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	@echo "Building for Linux (amd64)..."
	@cd $(SRC_DIR) && GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_AMD64) $(GO) build -o ../$(DIST_DIR)/$(BINARY_NAME)_$(GOOS_LINUX)_$(GOARCH_AMD64) -ldflags "-X main.version=$(VERSION)"
	@cd $(DIST_DIR) && zip -j $(BINARY_NAME)_$(GOOS_LINUX)_$(GOARCH_AMD64).zip $(BINARY_NAME)_$(GOOS_LINUX)_$(GOARCH_AMD64)
	
	@echo "Building for Windows (amd64)..."
	@cd $(SRC_DIR) && GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH_AMD64) $(GO) build -o ../$(DIST_DIR)/$(BINARY_NAME)_$(GOOS_WINDOWS)_$(GOARCH_AMD64).exe -ldflags "-X main.version=$(VERSION)"
	@cd $(DIST_DIR) && zip -j $(BINARY_NAME)_$(GOOS_WINDOWS)_$(GOARCH_AMD64).zip $(BINARY_NAME)_$(GOOS_WINDOWS)_$(GOARCH_AMD64).exe
	
	@echo "Building for macOS (amd64)..."
	@cd $(SRC_DIR) && GOOS=$(GOOS_DARWIN) GOARCH=$(GOARCH_AMD64) $(GO) build -o ../$(DIST_DIR)/$(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_AMD64) -ldflags "-X main.version=$(VERSION)"
	@cd $(DIST_DIR) && zip -j $(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_AMD64).zip $(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_AMD64)
	
	@echo "Building for macOS (arm64)..."
	@cd $(SRC_DIR) && GOOS=$(GOOS_DARWIN) GOARCH=$(GOARCH_ARM64) $(GO) build -o ../$(DIST_DIR)/$(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_ARM64) -ldflags "-X main.version=$(VERSION)"
	@cd $(DIST_DIR) && zip -j $(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_ARM64).zip $(BINARY_NAME)_$(GOOS_DARWIN)_$(GOARCH_ARM64)

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BIN_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Version management targets
.PHONY: version
version:
	@echo "$$(cat $(VERSION_FILE))"

.PHONY: bump-major
bump-major:
	@echo "Bumping major version..."
	@NEW_VERSION=$$(.github/scripts/increment_version.sh -t major -f $(VERSION_FILE)); \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	echo "New version: $$NEW_VERSION"; \
	if [ "$(shell uname)" = "Darwin" ]; then \
		sed -i '' -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	else \
		sed -i -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	fi
	@echo "Updated version in main.go"

.PHONY: bump-minor
bump-minor:
	@echo "Bumping minor version..."
	@NEW_VERSION=$$(.github/scripts/increment_version.sh -t minor -f $(VERSION_FILE)); \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	echo "New version: $$NEW_VERSION"; \
	if [ "$(shell uname)" = "Darwin" ]; then \
		sed -i '' -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	else \
		sed -i -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	fi
	@echo "Updated version in main.go"

.PHONY: bump-patch
bump-patch:
	@echo "Bumping patch version..."
	@NEW_VERSION=$$(.github/scripts/increment_version.sh -t patch -f $(VERSION_FILE)); \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	echo "New version: $$NEW_VERSION"; \
	if [ "$(shell uname)" = "Darwin" ]; then \
		sed -i '' -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	else \
		sed -i -E "s/(releaseVersion = \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1$$NEW_VERSION\2/g" ./src/main.go; \
	fi
	@echo "Updated version in main.go"

# Changelog targets
.PHONY: changelog
changelog:
	@echo "Generating changelog..."
	@.github/scripts/generate_changelog.sh --mode GENERATE --version $$(cat $(VERSION_FILE))
	@echo "Changelog generated in $(CHANGELOG_FILE)"

.PHONY: release-changelog
release-changelog:
	@echo "Generating changelog..."
	@.github/scripts/generate_changelog.sh --mode RELEASE --repo $(REPO_NAME) --version $$(cat $(VERSION_FILE)) --output-to-file
	@echo "Changelog generated in $(CHANGELOG_FILE)"

.PHONY: release-notes
release-notes:
	@echo "Generating release notes..."
	@.github/scripts/generate_changelog.sh --mode RELEASE_NOTES --version $$(cat $(VERSION_FILE))
	@echo "Release notes generated in $(RELEASE_NOTES_FILE)"

.PHONY: release
release: bump-patch changelog release-notes
	@echo "Release v$$(cat $(VERSION_FILE)) prepared"
	@echo "Changelog and release notes generated"
	@echo "Run 'git add $(VERSION_FILE) $(CHANGELOG_FILE) $(RELEASE_NOTES_FILE) src/main.go' to stage the changes"
	@echo "Then commit with 'git commit -m \"Release v$$(cat $(VERSION_FILE))\"'"

# Generate API documentation
.PHONY: docs
docs: check-tools
	@echo "Generating API documentation with Swag..."
	@cd $(SRC_DIR) && swag init -g main.go -o ../docs/swagger
	@echo "API documentation generated in docs/swagger"

# Run security scan with gosec
.PHONY: security-scan
security-scan: check-tools
	@echo "Running security scan with gosec..."
	@mkdir -p $(COVERAGE_DIR)
	@cd $(SRC_DIR) && gosec -no-fail -fmt=json -out=../$(COVERAGE_DIR)/security-scan.json ./...
	@cd $(SRC_DIR) && gosec -no-fail -fmt=html -out=../$(COVERAGE_DIR)/security-scan.html ./...
	@cd $(SRC_DIR) && gosec -no-fail -fmt=sarif -out=../$(COVERAGE_DIR)/security-scan.sarif ./...
	@echo "Security scan complete. Reports generated in $(COVERAGE_DIR): security-scan.json, security-scan.html, security-scan.sarif"

# Open security scan report in browser
.PHONY: security-report
security-report: security-scan
	@echo "Opening security scan report in browser..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open $(COVERAGE_DIR)/security-scan.html; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		xdg-open $(COVERAGE_DIR)/security-scan.html 2>/dev/null || echo "Please open $(COVERAGE_DIR)/security-scan.html in your browser"; \
	else \
		start $(COVERAGE_DIR)/security-scan.html 2>/dev/null || echo "Please open $(COVERAGE_DIR)/security-scan.html in your browser"; \
	fi

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all            - Run check-version, check-tools, clean, lint, test, build, and docs"
	@echo "  build          - Build the binary"
	@echo "  clean          - Remove build artifacts"
	@echo "  lint           - Run linting"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  coverage-report- Run tests with coverage and open the HTML report in a browser"
	@echo "  docs           - Generate API documentation with Swag"
	@echo "  setup          - Set up development environment (install required tools)"
	@echo "  check-tools    - Check if required tools are installed"
	@echo "  install-tools  - Install required Go tools"
	@echo "  check-version  - Check if version is set"
	@echo "  bump-major     - Bump major version (X.0.0)"
	@echo "  bump-minor     - Bump minor version (0.X.0)"
	@echo "  bump-patch     - Bump patch version (0.0.X)"
	@echo "  changelog      - Generate changelog"
	@echo "  release        - Create a new release"
	@echo "  security-scan  - Run security scan with gosec"
	@echo "  security-report- Open security scan report in browser"

# Setup development environment
.PHONY: setup
setup: check-version install-tools
	@echo "Development environment setup complete" 