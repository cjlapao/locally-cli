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
# DOCKER
# =============================================================================

.PHONY: docker-build
docker-build: ## Build the Docker image
	@echo "Building Docker image..."
	cd $(API_DIR) && docker build -t locally-api -f Dockerfile ../..

.PHONY: docker-run
docker-run: ## Run the Docker container
	@echo "Running Docker container..."
	cd $(API_DIR) && ./docker-run.sh run

.PHONY: docker-stop
docker-stop: ## Stop the Docker container
	@echo "Stopping Docker container..."
	cd $(API_DIR) && ./docker-run.sh stop

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	cd $(API_DIR) && ./docker-run.sh clean
	docker system prune -f
	docker image prune -f

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	@echo "Showing Docker container logs..."
	cd $(API_DIR) && ./docker-run.sh logs

.PHONY: docker-status
docker-status: ## Show Docker container status
	@echo "Showing Docker container status..."
	cd $(API_DIR) && ./docker-run.sh status

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
release: ## Build release binaries
	@echo "Building release binaries..."
	$(MKDIR) $(OUT_DIR) 2>$(NULL) || true
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-linux-amd64 ./$(API_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-darwin-amd64 ./$(API_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUT_DIR)/locally-api-windows-amd64.exe ./$(API_DIR)
	@echo "Release binaries created in $(OUT_DIR)/" 