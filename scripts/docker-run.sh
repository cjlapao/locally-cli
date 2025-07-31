#!/bin/bash

# Docker run helper script for Locally API (Linux/macOS)
# This script handles docker-compose issues and provides a more robust way to run the API

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
API_DIR="$PROJECT_ROOT/cmd/api"

# Docker registry configuration (can be overridden)
DOCKER_REGISTRY="${DOCKER_REGISTRY:-dcr.carloslapao.com}"
DOCKER_NAMESPACE="${DOCKER_NAMESPACE:-locally}"
DOCKER_IMAGE="${DOCKER_IMAGE:-locally-api}"
DOCKER_TAG="${DOCKER_TAG:-latest}"
DOCKER_FULL_NAME="${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${DOCKER_IMAGE}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if container exists
container_exists() {
    docker ps -a --format "table {{.Names}}" | grep -q "^locally-api$"
}

# Function to check if container is running
container_running() {
    docker ps --format "table {{.Names}}" | grep -q "^locally-api$"
}

# Function to get docker-compose version
get_docker_compose_version() {
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose --version 2>/dev/null | grep -oP 'docker-compose version \K[\d.]+' || echo ""
    else
        echo ""
    fi
}

# Function to run container with direct Docker commands
run_container_direct() {
    print_status "Using direct Docker commands (bypassing docker-compose)"
    
    # Build the image
    print_status "Building Docker image..."
    docker build -f "$API_DIR/Dockerfile" -t locally-api:latest "$PROJECT_ROOT"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to build Docker image"
        exit 1
    fi
    
    # Create network if it doesn't exist
    if ! docker network ls --format "table {{.Name}}" | grep -q "^locally-network$"; then
        print_status "Creating Docker network..."
        docker network create locally-network
    fi
    
    # Create volume if it doesn't exist
    if ! docker volume ls --format "table {{.Name}}" | grep -q "^locally-data$"; then
        print_status "Creating Docker volume..."
        docker volume create locally-data
    fi
    
    # Run the container
    print_status "Starting container..."
    docker run -d --name locally-api \
        --network locally-network \
        -p 8080:8080 \
        -v locally-data:/app/data \
        -e LOCALLY_DATABASE_TYPE=sqlite \
        -e LOCALLY_DATABASE_STORAGE_PATH=/app/data/locally.db \
        -e LOCALLY_SERVER_BIND_TO=0.0.0.0 \
        -e LOCALLY_SERVER_API_PORT=8080 \
        -e LOCALLY_JWT_AUTH_SECRET=your-jwt-secret-here-change-in-production \
        -e LOCALLY_ENCRYPTION_MASTER_SECRET=your-encryption-master-secret-change-in-production \
        -e LOCALLY_ENCRYPTION_GLOBAL_SECRET=your-encryption-global-secret-change-in-production \
        -e LOCALLY_MESSAGE_PROCESSOR_POLL_INTERVAL=10s \
        -e LOCALLY_MESSAGE_PROCESSOR_PROCESSING_TIMEOUT=30m \
        -e LOCALLY_MESSAGE_PROCESSOR_DEFAULT_MAX_RETRIES=3 \
        -e LOCALLY_MESSAGE_PROCESSOR_RECOVERY_ENABLED=true \
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_ENABLED=true \
        -e LOCALLY_MESSAGE_PROCESSOR_MAX_PROCESSING_AGE=1h \
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_MAX_AGE=24h \
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_INTERVAL=1h \
        -e LOCALLY_MESSAGE_PROCESSOR_KEEP_COMPLETE_MESSAGES=true \
        -e LOCALLY_MESSAGE_PROCESSOR_DEBUG=false \
        ${DOCKER_FULL_NAME}:${DOCKER_TAG}
    
    if [ $? -eq 0 ]; then
        print_success "Container started successfully"
        print_status "API will be available at http://localhost:8080"
        print_status "Use 'docker logs locally-api' to view logs"
    else
        print_error "Failed to start container"
        exit 1
    fi
}

# Function to stop container
stop_container() {
    if container_exists; then
        print_status "Stopping container..."
        docker stop locally-api
        if [ $? -eq 0 ]; then
            print_success "Container stopped"
        else
            print_error "Failed to stop container"
        fi
    else
        print_warning "Container 'locally-api' does not exist"
    fi
}

# Function to remove container
remove_container() {
    if container_exists; then
        print_status "Removing container..."
        docker rm -f locally-api
        if [ $? -eq 0 ]; then
            print_success "Container removed"
        else
            print_error "Failed to remove container"
        fi
    else
        print_warning "Container 'locally-api' does not exist"
    fi
}

# Function to show container logs
show_container_logs() {
    if container_exists; then
        print_status "Showing container logs..."
        docker logs locally-api
    else
        print_warning "Container 'locally-api' does not exist"
    fi
}

# Function to show container status
show_container_status() {
    print_status "Container status:"
    if container_exists; then
        if container_running; then
            print_success "Container 'locally-api' is running"
        else
            print_warning "Container 'locally-api' exists but is not running"
        fi
        docker ps -a --filter "name=locally-api" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    else
        print_warning "Container 'locally-api' does not exist"
    fi
}

# Function to clean up everything
clean_container() {
    print_status "Cleaning up containers, networks, and volumes..."
    
    # Stop and remove container
    if container_exists; then
        docker stop locally-api 2>/dev/null || true
        docker rm -f locally-api 2>/dev/null || true
    fi
    
    # Remove network
    docker network rm locally-network 2>/dev/null || true
    
    # Remove volume (WARNING: This will delete all data!)
    echo -n "Do you want to delete the database volume? This will delete all data! (y/N): "
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        docker volume rm locally-data 2>/dev/null || true
        print_success "Database volume removed"
    else
        print_status "Database volume preserved"
    fi
    
    print_success "Cleanup completed"
}

# Main execution
print_status "Docker Run Script for Locally API (Linux/macOS)"

ACTION="${1:-run}"

case "$ACTION" in
    "run")
        # Check if container is already running
        if container_running; then
            print_warning "Container 'locally-api' is already running"
            print_status "Use 'stop' to stop it first, or 'restart' to restart it"
            exit 0
        fi
        
        # Check if container exists but is stopped
        if container_exists; then
            print_status "Container exists but is stopped. Starting it..."
            docker start locally-api
            if [ $? -eq 0 ]; then
                print_success "Container started successfully"
            else
                print_error "Failed to start existing container"
                exit 1
            fi
        else
            # Check docker-compose version
            COMPOSE_VERSION=$(get_docker_compose_version)
            if [ -n "$COMPOSE_VERSION" ] && [ "$(echo "$COMPOSE_VERSION" | cut -d. -f1)" -lt 2 ]; then
                print_warning "Detected docker-compose version $COMPOSE_VERSION (known to have issues)"
                print_status "Using direct Docker commands instead"
                run_container_direct
            else
                print_status "Using docker-compose..."
                cd "$API_DIR"
                # Export build arguments for docker-compose
                export VERSION="${VERSION:-0.0.0}"
                export BUILD_TIME="${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}"
                export GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
                if docker-compose up -d --build; then
                    print_success "Container started successfully with docker-compose"
                else
                    print_warning "docker-compose failed, trying direct Docker commands"
                    run_container_direct
                fi
            fi
        fi
        ;;
    
    "stop")
        stop_container
        ;;
    
    "restart")
        stop_container
        sleep 2
        "$0" "run"
        ;;
    
    "logs")
        show_container_logs
        ;;
    
    "status")
        show_container_status
        ;;
    
    "clean")
        clean_container
        ;;
    
    *)
        print_error "Unknown action: $ACTION"
        print_status "Available actions: run, stop, restart, logs, status, clean"
        exit 1
        ;;
esac 