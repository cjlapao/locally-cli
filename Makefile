# Locally CLI Makefile
# Cross-platform compatible (Linux/macOS/Windows)

# Target selection (default: api)
TARGET ?= api

# Version variables
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags for version injection
LDFLAGS = -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

# Validate target
ifeq ($(filter api cli,$(TARGET)),)
	$(error Invalid TARGET: $(TARGET). Valid targets are: api, cli)
endif

# Variables based on target
ifeq ($(TARGET),api)
	CMD_DIR = cmd/api
	BINARY_NAME = locally-api
	ENV_FILE = $(CMD_DIR)/.api.env
	ENV_TEMPLATE = $(CMD_DIR)/env.template
	HAS_DOCKER = true
else
	CMD_DIR = cmd/cli
	BINARY_NAME = locally-cli
	ENV_FILE = $(CMD_DIR)/.cli.env
	ENV_TEMPLATE = 
	HAS_DOCKER = false
endif

OUT_DIR = out

# Detect OS for cross-platform compatibility
ifeq ($(OS),Windows_NT)
	# Windows
	RM = del /Q
	MKDIR = mkdir
	CP = copy
	RMDIR = rmdir /S /Q
	SEP = \\
	NULL = nul
	EXIST_CHECK = if exist
	NOT_EXIST_CHECK = if not exist
	COPY_CMD = copy
	MKDIR_CMD = mkdir
else
	# Linux/macOS
	RM = rm -f
	MKDIR = mkdir -p
	CP = cp
	RMDIR = rm -rf
	SEP = /
	NULL = /dev/null
	EXIST_CHECK = test -f
	NOT_EXIST_CHECK = test ! -f
	COPY_CMD = cp
	MKDIR_CMD = mkdir -p
endif

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Locally CLI - Available Commands (Target: $(TARGET))"
	@echo ""
	@echo "  TARGET SELECTION:"
	@echo "    TARGET=api        Build for API (default)"
	@echo "    TARGET=cli        Build for CLI"
	@echo "    Examples:"
	@echo "      make build              # Build API (default)"
	@echo "      make TARGET=cli build  # Build CLI"
	@echo "      make TARGET=api build  # Build API explicitly"
	@echo ""
	@echo "  INITIALIZATION:"
	@echo "    init              Initialize the project (create .api.env, out folder, and run go mod)"
	@echo "    setup             Complete project setup (init + deps)"
	@echo ""
	@echo "  BUILDING:"
	@echo "    build             Build the $(BINARY_NAME) binary"
	@echo "    build-cross       Build for specific platform (GOOS=linux GOARCH=amd64 make build-cross)"
	@echo "    build-linux       Build for Linux (amd64)"
	@echo "    build-macos       Build for macOS (amd64)"
	@echo "    build-windows     Build for Windows (amd64)"
	@echo "    build-all         Build for all major platforms"
	@echo "    release           Build release binaries for all platforms"
	@echo ""
	@echo "  DEVELOPMENT:"
ifeq ($(TARGET),api)
	@echo "    dev               Run the API in development mode"
endif
	@echo "    api-build         Build only the API service"
	@echo "    cli-build         Build only the CLI service"
	@echo ""
ifeq ($(TARGET),api)
	@echo "  DOCKER (API only):"
	@echo "    docker-build      Build the Docker image"
	@echo "    docker-run        Run the Docker container"
	@echo "    docker-stop       Stop the Docker container"
	@echo "    docker-clean      Clean Docker images and containers"
	@echo "    docker-logs       Show Docker container logs"
	@echo "    docker-status     Show Docker container status"
	@echo ""
	@echo "  DOCKER REGISTRY (API only):"
	@echo "    docker-login      Login to Docker registry"
	@echo "    docker-push       Build and push Docker image to registry (both latest and version tags)"
	@echo "    docker-push-latest Build and push with latest tag only"
	@echo "    docker-push-version Build and push with version tag only"
	@echo "    docker-push-beta  Build and push with beta tag (timestamp format)"
	@echo "    docker-pull       Pull Docker image from registry"
	@echo "    docker-tag        Tag local image for registry"
	@echo ""
else
	@echo "  DOCKER: Not available for CLI target"
	@echo ""
endif
	@echo "  TESTING:"
	@echo "    test              Run all tests"
	@echo "    test-verbose      Run tests with verbose output"
	@echo "    test-coverage     Run tests with coverage report"
	@echo ""
	@echo "  UTILITIES:"
	@echo "    clean             Clean build artifacts and temporary files"
	@echo "    deps              Download and tidy dependencies"
	@echo "    update-deps       Update dependencies to latest versions"
	@echo "    check-env         Check if environment file exists"
	@echo "    version           Show current version"
	@echo "    beta-version      Show current beta version (timestamp)"
	@echo "    fmt               Format Go code"
	@echo "    vet               Run go vet"
	@echo "    lint              Run all linting checks"
	@echo ""
	@echo "  ALL-IN-ONE:"
	@echo "    full-build        Full build with linting and testing"
ifeq ($(TARGET),api)
	@echo "    docker-full       Build and run Docker container"
endif
	@echo ""
	@echo "For detailed information about each command, see the README.md file."

# =============================================================================
# INITIALIZATION
# =============================================================================

.PHONY: init
init: ## Initialize the project (create .api.env, out folder, and run go mod)
	@echo "Initializing Locally CLI project for $(TARGET)..."
	@echo "Creating out directory..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
ifeq ($(TARGET),api)
	@echo "Checking for .api.env file..."
ifeq ($(OS),Windows_NT)
	@if not exist "$(ENV_FILE)" ( \
		echo "Creating .api.env from template..." && \
		$(CP) "$(ENV_TEMPLATE)" "$(ENV_FILE)" && \
		echo "Created $(ENV_FILE) - please edit with your configuration" \
	) else ( \
		echo "$(ENV_FILE) already exists - skipping" \
	)
else
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "Creating .api.env from template..."; \
		$(CP) "$(ENV_TEMPLATE)" "$(ENV_FILE)"; \
		echo "Created $(ENV_FILE) - please edit with your configuration"; \
	else \
		echo "$(ENV_FILE) already exists - skipping"; \
	fi
endif
else
	@echo "Checking for .cli.env file..."
ifeq ($(OS),Windows_NT)
	@if not exist "$(ENV_FILE)" ( \
		echo "Creating .cli.env file..." && \
		echo "# Locally CLI Environment Variables" > "$(ENV_FILE)" && \
		echo "# VS Code will automatically recognize this file" >> "$(ENV_FILE)" && \
		echo "" >> "$(ENV_FILE)" && \
		echo "# Debug mode" >> "$(ENV_FILE)" && \
		echo "LOCALLY_DEBUG=false" >> "$(ENV_FILE)" && \
		echo "" >> "$(ENV_FILE)" && \
		echo "# API endpoint" >> "$(ENV_FILE)" && \
		echo "LOCALLY_API_ENDPOINT=http://localhost:8080" >> "$(ENV_FILE)" && \
		echo "Created $(ENV_FILE) - please edit with your configuration" \
	) else ( \
		echo "$(ENV_FILE) already exists - skipping" \
	)
else
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "Creating .cli.env file..."; \
		echo "# Locally CLI Environment Variables" > "$(ENV_FILE)"; \
		echo "# VS Code will automatically recognize this file" >> "$(ENV_FILE)"; \
		echo "" >> "$(ENV_FILE)"; \
		echo "# Debug mode" >> "$(ENV_FILE)"; \
		echo "LOCALLY_DEBUG=false" >> "$(ENV_FILE)"; \
		echo "" >> "$(ENV_FILE)"; \
		echo "# API endpoint" >> "$(ENV_FILE)"; \
		echo "LOCALLY_API_ENDPOINT=http://localhost:8080" >> "$(ENV_FILE)"; \
		echo "Created $(ENV_FILE) - please edit with your configuration"; \
	else \
		echo "$(ENV_FILE) already exists - skipping"; \
	fi
endif
endif
	@echo "Running go mod tidy (this may show warnings for missing packages)..."
	-go mod tidy -e
	@echo "Running go mod download..."
	-go mod download
	@echo "Testing $(TARGET) build..."
	@echo "Building $(TARGET) only..."
	-go build -ldflags "$(LDFLAGS)" -o /tmp/test-build ./$(CMD_DIR) 2>/dev/null || echo "Warning: $(TARGET) build failed due to missing dependencies"
	@echo "Initialization complete!"
	@echo ""
	@echo "Next steps:"
ifeq ($(TARGET),api)
	@echo "1. Edit $(ENV_FILE) with your configuration"
	@echo "2. Run 'make dev' to start the API in development mode"
	@echo "3. Run 'make docker-run' to start with Docker"
else
	@echo "1. Run 'make build' to build the CLI"
	@echo "2. Run 'make test' to run tests"
endif

# =============================================================================
# CLEANUP
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@echo "Cleaning build artifacts..."
	$(RMDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Cleaning Go cache..."
	go clean -cache -modcache -testcache
	@echo "Cleanup complete!"

# =============================================================================
# DEVELOPMENT
# =============================================================================

.PHONY: dev
dev: ## Run the API in development mode
ifeq ($(TARGET),api)
	@echo "Starting API in development mode..."
	cd $(CMD_DIR) && go run main.go
else
	@echo "Error: 'dev' target is only available for API (TARGET=api)"
	@echo "Current target: $(TARGET)"
	@exit 1
endif

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME) with version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Building $(TARGET) service..."
	-go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR) || (echo "Build failed. Trying to build with missing dependencies..." && go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR) 2>/dev/null || echo "Build failed due to missing dependencies. Please check the codebase.")
	@echo "Build complete!"

.PHONY: api-build
api-build: ## Build only the API service (ignores other packages)
	@echo "Building API service only with version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Building API binary..."
	cd cmd/api && go build -ldflags "$(LDFLAGS)" -o ../../$(OUT_DIR)/locally-api .
	@echo "API build complete!"

.PHONY: cli-build
cli-build: ## Build only the CLI service (ignores other packages)
	@echo "Building CLI service only with version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Building CLI binary..."
	cd cmd/cli && go build -ldflags "$(LDFLAGS)" -o ../../$(OUT_DIR)/locally-cli .
	@echo "CLI build complete!"

# =============================================================================
# CROSS-PLATFORM BUILDING
# =============================================================================

# Build parameters (can be overridden)
GOOS ?= $(shell go env GOOS 2>/dev/null || echo linux)
GOARCH ?= $(shell go env GOARCH 2>/dev/null || echo amd64)
CGO_ENABLED ?= 0

.PHONY: build-cross
build-cross: ## Build for specific platform (GOOS=linux GOARCH=amd64 make build-cross)
	@echo "Building for $(GOOS)/$(GOARCH) with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
ifeq ($(GOOS),windows)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe"
else
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)"
endif

.PHONY: build-linux
build-linux: ## Build for Linux (amd64)
	@echo "Building for linux/amd64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-linux-arm64
build-linux-arm64: ## Build for Linux (arm64)
	@echo "Building for linux/arm64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-linux-arm64"

.PHONY: build-macos
build-macos: ## Build for macOS (amd64)
	@echo "Building for darwin/amd64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64"

.PHONY: build-macos-arm64
build-macos-arm64: ## Build for macOS (arm64/M1)
	@echo "Building for darwin/arm64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64"

.PHONY: build-windows
build-windows: ## Build for Windows (amd64)
	@echo "Building for windows/amd64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe"

.PHONY: build-windows-arm64
build-windows-arm64: ## Build for Windows (arm64)
	@echo "Building for windows/arm64 with CGO_ENABLED=$(CGO_ENABLED) and version $(VERSION)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-windows-arm64.exe ./$(CMD_DIR)
	@echo "Build complete: $(OUT_DIR)/$(BINARY_NAME)-windows-arm64.exe"

.PHONY: build-all
build-all: ## Build for all major platforms (current platform + same OS alternatives)
	@echo "Building for current platform and same OS alternatives..."
	@make build-linux
	@make build-macos
	@echo "Note: Cross-platform builds require target platform C compilers"
	@echo "For Windows builds, build natively on Windows or use CI/CD"
	@echo "All available platform builds complete!"



# =============================================================================
# DOCKER
# =============================================================================

# Docker registry configuration
DOCKER_REGISTRY = dcr.carloslapao.com
DOCKER_NAMESPACE = locally
DOCKER_IMAGE = locally-api
DOCKER_TAG ?= latest
DOCKER_FULL_NAME = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE)

# Version is already defined at the top of the file

# Generate beta version with timestamp (ddmmyyhhMM format)
BETA_VERSION ?= $(shell date +%d%m%y%H%M)

# Docker target validation
.PHONY: validate-docker
validate-docker:
ifeq ($(HAS_DOCKER),false)
	@echo "Error: Docker targets are not available for $(TARGET) target"
	@echo "Docker is only available for API (TARGET=api)"
	@echo "Current target: $(TARGET)"
	@exit 1
endif

.PHONY: docker-build
docker-build: validate-docker ## Build the Docker image
	@echo "Building Docker image..."
	cd $(API_DIR) && docker build -t $(DOCKER_FULL_NAME):$(DOCKER_TAG) -f Dockerfile ../..

.PHONY: docker-run
docker-run: validate-docker ## Run the Docker container
	@echo "Running Docker container..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 run
else
	./scripts/docker-run.sh run
endif

.PHONY: docker-stop
docker-stop: validate-docker ## Stop the Docker container
	@echo "Stopping Docker container..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 stop
else
	./scripts/docker-run.sh stop
endif

.PHONY: docker-clean
docker-clean: validate-docker ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 clean
else
	./scripts/docker-run.sh clean
endif
	docker system prune -f
	docker image prune -f

.PHONY: docker-logs
docker-logs: validate-docker ## Show Docker container logs
	@echo "Showing Docker container logs..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 logs
else
	./scripts/docker-run.sh logs
endif

.PHONY: docker-status
docker-status: validate-docker ## Show Docker container status
	@echo "Showing Docker container status..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 status
else
	./scripts/docker-run.sh status
endif

# =============================================================================
# DOCKER REGISTRY
# =============================================================================

.PHONY: docker-login
docker-login: validate-docker ## Login to Docker registry
	@echo "Logging in to $(DOCKER_REGISTRY)..."
	docker login $(DOCKER_REGISTRY)

.PHONY: docker-push
docker-push: validate-docker docker-build ## Build and push Docker image to registry (both latest and version tags)
	@echo "Building and pushing with both latest and version tags..."
	@echo "Tagging with latest..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):latest
	@echo "Tagging with version $(VERSION)..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):$(VERSION)
	@echo "Pushing latest tag..."
	docker push $(DOCKER_FULL_NAME):latest
	@echo "Pushing version tag $(VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(VERSION)
	@echo "Successfully pushed both latest and $(VERSION) tags"

.PHONY: docker-push-latest
docker-push-latest: validate-docker ## Build and push with latest tag only
	@echo "Building and pushing with latest tag only..."
	@make docker-build
	@echo "Pushing latest tag..."
	docker push $(DOCKER_FULL_NAME):latest

.PHONY: docker-push-version
docker-push-version: validate-docker ## Build and push with version tag only
	@echo "Building and pushing with version tag only..."
	@make docker-build
	@echo "Pushing version tag $(VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(VERSION)

.PHONY: docker-push-beta
docker-push-beta: validate-docker ## Build and push with beta tag (timestamp format)
	@echo "Building and pushing with beta tag..."
	@make docker-build
	@echo "Tagging with beta version $(BETA_VERSION)..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):$(BETA_VERSION)
	@echo "Pushing beta tag $(BETA_VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(BETA_VERSION)

.PHONY: docker-pull
docker-pull: validate-docker ## Pull Docker image from registry
	@echo "Pulling $(DOCKER_FULL_NAME):$(DOCKER_TAG) from registry..."
	docker pull $(DOCKER_FULL_NAME):$(DOCKER_TAG)

.PHONY: docker-tag
docker-tag: validate-docker ## Tag local image with registry name
	@echo "Tagging locally-api as $(DOCKER_FULL_NAME):$(DOCKER_TAG)..."
	docker tag locally-api $(DOCKER_FULL_NAME):$(DOCKER_TAG)

.PHONY: docker-build-and-push
docker-build-and-push: docker-build ## Build, tag, and push in one command (both latest and version tags)
	@echo "Building and pushing with both latest and version tags..."
	@echo "Tagging with latest..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):latest
	@echo "Tagging with version $(VERSION)..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):$(VERSION)
	@echo "Pushing latest tag..."
	docker push $(DOCKER_FULL_NAME):latest
	@echo "Pushing version tag $(VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(VERSION)
	@echo "Successfully pushed both latest and $(VERSION) tags"

.PHONY: docker-build-and-push-latest
docker-build-and-push-latest: ## Build and push with latest tag only
	@echo "Building and pushing with latest tag only..."
	@make docker-build
	@echo "Pushing latest tag..."
	docker push $(DOCKER_FULL_NAME):latest

.PHONY: docker-build-and-push-version
docker-build-and-push-version: ## Build and push with version tag only
	@echo "Building and pushing with version tag only..."
	@make docker-build
	@echo "Pushing version tag $(VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(VERSION)

.PHONY: docker-build-and-push-beta
docker-build-and-push-beta: ## Build and push with beta tag (timestamp format)
	@echo "Building and pushing with beta tag..."
	@make docker-build
	@echo "Tagging with beta version $(BETA_VERSION)..."
	docker tag $(DOCKER_FULL_NAME):$(DOCKER_TAG) $(DOCKER_FULL_NAME):$(BETA_VERSION)
	@echo "Pushing beta tag $(BETA_VERSION)..."
	docker push $(DOCKER_FULL_NAME):$(BETA_VERSION)

.PHONY: version
version: ## Show current version
	@echo "Current version: $(VERSION)"

.PHONY: beta-version
beta-version: ## Show current beta version (timestamp)
	@echo "Current beta version: $(BETA_VERSION)"

# =============================================================================
# TESTING
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	go test -coverprofile=$(OUT_DIR)/coverage.out ./...
	go tool cover -html=$(OUT_DIR)/coverage.out -o $(OUT_DIR)/coverage.html
	@echo "Coverage report generated: $(OUT_DIR)/coverage.html"

# =============================================================================
# LINTING & FORMATTING
# =============================================================================

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: fmt vet ## Run all linting checks

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy -e

.PHONY: update-deps
update-deps: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy -e

.PHONY: check-env
check-env: ## Check if environment file exists
	@echo "Checking environment configuration..."
ifeq ($(OS),Windows_NT)
	@if exist "$(ENV_FILE)" ( \
		echo "✓ $(ENV_FILE) exists" \
	) else ( \
		echo "✗ $(ENV_FILE) not found - run 'make init' to create it" \
	)
else
	@if [ -f "$(ENV_FILE)" ]; then \
		echo "✓ $(ENV_FILE) exists"; \
	else \
		echo "✗ $(ENV_FILE) not found - run 'make init' to create it"; \
	fi
endif

# =============================================================================
# ALL-IN-ONE COMMANDS
# =============================================================================

.PHONY: setup
setup: init deps ## Complete project setup (init + deps)

.PHONY: full-build
full-build: lint test build ## Full build with linting and testing

.PHONY: docker-full
docker-full: validate-docker docker-build docker-run ## Build and run Docker container

# =============================================================================
# WINDOWS COMPATIBILITY HELPERS
# =============================================================================

.PHONY: windows-init
windows-init: ## Windows-specific initialization (alternative to init)
	@echo "Windows-specific initialization..."
	@if not exist "out" mkdir out
	@if not exist "$(API_DIR)\.api.env" ( \
		copy "$(API_DIR)\env.template" "$(API_DIR)\.api.env" && \
		echo "Created $(API_DIR)\.api.env - please edit with your configuration" \
	) else ( \
		echo "$(API_DIR)\.api.env already exists - skipping" \
	)
	go mod tidy -e
	go mod download
	@echo "Windows initialization complete!"

# =============================================================================
# DOCUMENTATION
# =============================================================================

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@echo "Documentation generation not implemented yet"

# =============================================================================
# RELEASE
# =============================================================================

.PHONY: release
release: ## Build release binaries for all platforms
	@echo "Building release binaries for all platforms..."
	@make build-all
	@echo "Release binaries created in $(OUT_DIR)/"

 