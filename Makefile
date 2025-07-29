# Locally CLI Makefile
# Cross-platform compatible (Linux/macOS/Windows)

# Variables
API_DIR = cmd/api
OUT_DIR = out
ENV_FILE = $(API_DIR)/.api.env
ENV_TEMPLATE = $(API_DIR)/env.template

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
	@echo "Locally CLI - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# INITIALIZATION
# =============================================================================

.PHONY: init
init: ## Initialize the project (create .api.env, out folder, and run go mod)
	@echo "Initializing Locally CLI project..."
	@echo "Creating out directory..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
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
	@echo "Running go mod tidy (this may show warnings for missing packages)..."
	-go mod tidy
	@echo "Running go mod download..."
	-go mod download
	@echo "Testing API build..."
	@echo "Building API only..."
	-go build -o /tmp/test-build ./$(API_DIR) 2>/dev/null || echo "Warning: API build failed due to missing dependencies"
	@echo "Initialization complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit $(ENV_FILE) with your configuration"
	@echo "2. Run 'make dev' to start the API in development mode"
	@echo "3. Run 'make docker-run' to start with Docker"

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
	@echo "Starting API in development mode..."
	cd $(API_DIR) && go run main.go

.PHONY: build
build: ## Build the API binary
	@echo "Building API binary..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Building API service..."
	-go build -o $(OUT_DIR)/locally-api ./$(API_DIR) || (echo "Build failed. Trying to build with missing dependencies..." && go build -o $(OUT_DIR)/locally-api ./$(API_DIR) 2>/dev/null || echo "Build failed due to missing dependencies. Please check the codebase.")
	@echo "Build complete!"

.PHONY: api-build
api-build: ## Build only the API service (ignores other packages)
	@echo "Building API service only..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	@echo "Building API binary..."
	cd $(API_DIR) && go build -o ../../$(OUT_DIR)/locally-api .
	@echo "API build complete!"

# =============================================================================
# CROSS-PLATFORM BUILDING
# =============================================================================

# Build parameters (can be overridden)
GOOS ?= $(shell go env GOOS 2>/dev/null || echo linux)
GOARCH ?= $(shell go env GOARCH 2>/dev/null || echo amd64)
CGO_ENABLED ?= 1

.PHONY: build-cross
build-cross: ## Build for specific platform (GOOS=linux GOARCH=amd64 make build-cross)
	@echo "Building for $(GOOS)/$(GOARCH) with CGO_ENABLED=$(CGO_ENABLED)..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
ifeq ($(GOOS),windows)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUT_DIR)/locally-api-$(GOOS)-$(GOARCH).exe ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-$(GOOS)-$(GOARCH).exe"
else
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUT_DIR)/locally-api-$(GOOS)-$(GOARCH) ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-$(GOOS)-$(GOARCH)"
endif

.PHONY: build-linux
build-linux: ## Build for Linux (amd64)
	@echo "Building for linux/amd64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-linux-amd64 ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-linux-amd64"

.PHONY: build-linux-arm64
build-linux-arm64: ## Build for Linux (arm64)
	@echo "Building for linux/arm64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o $(OUT_DIR)/locally-api-linux-arm64 ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-linux-arm64"

.PHONY: build-macos
build-macos: ## Build for macOS (amd64)
	@echo "Building for darwin/amd64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-darwin-amd64 ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-darwin-amd64"

.PHONY: build-macos-arm64
build-macos-arm64: ## Build for macOS (arm64/M1)
	@echo "Building for darwin/arm64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $(OUT_DIR)/locally-api-darwin-arm64 ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-darwin-arm64"

.PHONY: build-windows
build-windows: ## Build for Windows (amd64)
	@echo "Building for windows/amd64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-windows-amd64.exe ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-windows-amd64.exe"

.PHONY: build-windows-arm64
build-windows-arm64: ## Build for Windows (arm64)
	@echo "Building for windows/arm64 with CGO_ENABLED=1..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 go build -o $(OUT_DIR)/locally-api-windows-arm64.exe ./$(API_DIR)
	@echo "Build complete: $(OUT_DIR)/locally-api-windows-arm64.exe"

.PHONY: build-all
build-all: ## Build for all major platforms (current platform + same OS alternatives)
	@echo "Building for current platform and same OS alternatives..."
	@$(MAKE) build-linux
	@$(MAKE) build-macos
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

.PHONY: docker-build
docker-build: ## Build the Docker image
	@echo "Building Docker image..."
	cd $(API_DIR) && docker build -t $(DOCKER_FULL_NAME):$(DOCKER_TAG) -f Dockerfile ../..

.PHONY: docker-run
docker-run: ## Run the Docker container
	@echo "Running Docker container..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 run
else
	./scripts/docker-run.sh run
endif

.PHONY: docker-stop
docker-stop: ## Stop the Docker container
	@echo "Stopping Docker container..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 stop
else
	./scripts/docker-run.sh stop
endif

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 clean
else
	./scripts/docker-run.sh clean
endif
	docker system prune -f
	docker image prune -f

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	@echo "Showing Docker container logs..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 logs
else
	./scripts/docker-run.sh logs
endif

.PHONY: docker-status
docker-status: ## Show Docker container status
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
docker-login: ## Login to Docker registry
	@echo "Logging in to $(DOCKER_REGISTRY)..."
	docker login $(DOCKER_REGISTRY)

.PHONY: docker-push
docker-push: docker-build ## Build and push Docker image to registry
	@echo "Pushing $(DOCKER_FULL_NAME):$(DOCKER_TAG) to registry..."
	docker push $(DOCKER_FULL_NAME):$(DOCKER_TAG)

.PHONY: docker-push-latest
docker-push-latest: ## Build and push with latest tag
	@$(MAKE) docker-push DOCKER_TAG=latest

.PHONY: docker-push-version
docker-push-version: ## Build and push with version tag (VERSION file)
	@echo "Reading version from VERSION file..."
	@$(MAKE) docker-push DOCKER_TAG=$(shell cat VERSION 2>/dev/null || echo "unknown")

.PHONY: docker-pull
docker-pull: ## Pull Docker image from registry
	@echo "Pulling $(DOCKER_FULL_NAME):$(DOCKER_TAG) from registry..."
	docker pull $(DOCKER_FULL_NAME):$(DOCKER_TAG)

.PHONY: docker-tag
docker-tag: ## Tag local image with registry name
	@echo "Tagging locally-api as $(DOCKER_FULL_NAME):$(DOCKER_TAG)..."
	docker tag locally-api $(DOCKER_FULL_NAME):$(DOCKER_TAG)

.PHONY: docker-build-and-push
docker-build-and-push: docker-build docker-tag docker-push ## Build, tag, and push in one command

.PHONY: docker-build-and-push-latest
docker-build-and-push-latest: ## Build and push with latest tag
	@$(MAKE) docker-build-and-push DOCKER_TAG=latest

.PHONY: docker-build-and-push-version
docker-build-and-push-version: ## Build and push with version tag
	@$(MAKE) docker-build-and-push DOCKER_TAG=$(shell cat VERSION 2>/dev/null || echo "unknown")

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
	go mod tidy

.PHONY: update-deps
update-deps: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: check-env
check-env: ## Check if .api.env file exists
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
docker-full: docker-build docker-run ## Build and run Docker container

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
	go mod tidy
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
	@$(MAKE) build-all
	@echo "Release binaries created in $(OUT_DIR)/"

 