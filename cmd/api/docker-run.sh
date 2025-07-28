#!/bin/bash

# Docker run helper script for Locally API
# This script handles docker-compose issues and provides a more robust way to run the API

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

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

# Function to stop and remove container
cleanup_container() {
    if container_exists; then
        print_status "Stopping existing container..."
        docker stop locally-api 2>/dev/null || true
        print_status "Removing existing container..."
        docker rm locally-api 2>/dev/null || true
        print_success "Container cleaned up"
    fi
}

# Function to check docker-compose version
check_docker_compose() {
    local version
    version=$(docker-compose --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    print_status "Docker Compose version: $version"
    
    # Check if version is known to have ContainerConfig issues
    if [[ "$version" == "1.29.2" ]]; then
        print_warning "Docker Compose $version is known to have ContainerConfig issues"
        print_warning "Using alternative approach for container management"
        return 1
    fi
    return 0
}

# Function to run with docker-compose
run_with_compose() {
    print_status "Starting with docker-compose..."
    cd "$SCRIPT_DIR"
    
    # Try to bring down any existing containers first
    docker-compose down 2>/dev/null || true
    
    # Start the service
    if docker-compose up -d --build; then
        print_success "Container started successfully with docker-compose"
        return 0
    else
        print_error "Docker-compose failed, trying alternative approach"
        return 1
    fi
}

# Function to run with direct docker commands
run_with_docker() {
    print_status "Starting with direct Docker commands..."
    
    # Build the image
    print_status "Building Docker image..."
    docker build -t locally-api:latest -f "$SCRIPT_DIR/Dockerfile" "$PROJECT_ROOT"
    
    # Create network if it doesn't exist
    if ! docker network ls --format "table {{.Name}}" | grep -q "^locally-network$"; then
        print_status "Creating network..."
        docker network create locally-network
    fi
    
    # Create volume if it doesn't exist
    if ! docker volume ls --format "table {{.Name}}" | grep -q "^locally-data$"; then
        print_status "Creating volume..."
        docker volume create locally-data
    fi
    
    # Run the container
    print_status "Starting container..."
    docker run -d \
        --name locally-api \
        --network locally-network \
        -p 8080:8080 \
        -v locally-data:/app/data \
        --restart unless-stopped \
        --env-file "$SCRIPT_DIR/env.template" \
        -e LOCALLY_DATABASE_TYPE=sqlite \
        -e LOCALLY_DATABASE_STORAGE_PATH=/app/data/locally.db \
        -e LOCALLY_SERVER_BIND_TO=0.0.0.0 \
        -e LOCALLY_SERVER_API_PORT=8080 \
        -e LOCALLY_JWT_AUTH_SECRET=your-jwt-secret-here-change-in-production \
        -e LOCALLY_ENCRYPTION_MASTER_SECRET=your-encryption-master-secret-change-in-production \
        -e LOCALLY_ENCRYPTION_GLOBAL_SECRET=your-encryption-global-secret-change-in-production \
        -e LOCALLY_AUTH_ROOT_PASSWORD=your-secure-root-password-change-in-production \
        -e LOCALLY_CORS_ALLOW_ORIGINS=* \
        -e LOCALLY_CORS_ALLOW_METHODS=OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE \
        -e LOCALLY_CORS_ALLOW_HEADERS=* \
        -e LOCALLY_LOG_LEVEL=info \
        -e LOCALLY_DEBUG=false \
        -e LOCALLY_DATABASE_MIGRATE=true \
        -e LOCALLY_MESSAGE_PROCESSOR_POLL_INTERVAL=10s \
        -e LOCALLY_MESSAGE_PROCESSOR_PROCESSING_TIMEOUT=30m \
        -e LOCALLY_MESSAGE_PROCESSOR_DEFAULT_MAX_RETRIES=3 \
        -e LOCALLY_MESSAGE_PROCESSOR_RECOVERY_ENABLED=true \
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_ENABLED=true \
        -e LOCALLY_SEED_DEMO_DATA=false \
        locally-api:latest
    
    print_success "Container started successfully with Docker"
}

# Function to show logs
show_logs() {
    print_status "Showing container logs..."
    docker logs -f locally-api
}

# Function to show status
show_status() {
    print_status "Container status:"
    if container_running; then
        print_success "Container is running"
        docker ps --filter "name=locally-api"
    elif container_exists; then
        print_warning "Container exists but is not running"
        docker ps -a --filter "name=locally-api"
    else
        print_warning "Container does not exist"
    fi
}

# Main execution
main() {
    print_status "Locally API Docker Runner"
    print_status "Script directory: $SCRIPT_DIR"
    print_status "Project root: $PROJECT_ROOT"
    
    # Parse command line arguments
    case "${1:-run}" in
        "run")
            # Check if container is already running
            if container_running; then
                print_warning "Container is already running"
                show_status
                exit 0
            fi
            
            # Clean up if container exists but is not running
            if container_exists; then
                cleanup_container
            fi
            
            # Try docker-compose first, fallback to direct docker
            if check_docker_compose && run_with_compose; then
                print_success "Started successfully with docker-compose"
            else
                run_with_docker
            fi
            
            # Wait a moment and show status
            sleep 2
            show_status
            ;;
        "stop")
            print_status "Stopping container..."
            docker stop locally-api 2>/dev/null || true
            print_success "Container stopped"
            ;;
        "clean")
            print_status "Cleaning up container and volumes..."
            docker stop locally-api 2>/dev/null || true
            docker rm locally-api 2>/dev/null || true
            docker volume rm locally-data 2>/dev/null || true
            docker network rm locally-network 2>/dev/null || true
            print_success "Cleanup completed"
            ;;
        "logs")
            show_logs
            ;;
        "status")
            show_status
            ;;
        "restart")
            print_status "Restarting container..."
            docker restart locally-api 2>/dev/null || {
                print_warning "Container not running, starting fresh..."
                cleanup_container
                main run
            }
            ;;
        *)
            echo "Usage: $0 {run|stop|clean|logs|status|restart}"
            echo "  run     - Start the container (default)"
            echo "  stop    - Stop the container"
            echo "  clean   - Stop and remove container, volumes, and networks"
            echo "  logs    - Show container logs"
            echo "  status  - Show container status"
            echo "  restart - Restart the container"
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 