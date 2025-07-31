# Docker run helper script for Locally API (Windows PowerShell)
# This script handles docker-compose issues and provides a more robust way to run the API

param(
    [Parameter(Position=0)]
    [ValidateSet("run", "stop", "restart", "logs", "status", "clean")]
    [string]$Action = "run"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Get script directory and project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ApiDir = Join-Path $ProjectRoot "cmd\api"

# Docker registry configuration (can be overridden)
$DockerRegistry = if ($env:DOCKER_REGISTRY) { $env:DOCKER_REGISTRY } else { "dcr.carloslapao.com" }
$DockerNamespace = if ($env:DOCKER_NAMESPACE) { $env:DOCKER_NAMESPACE } else { "locally" }
$DockerImage = if ($env:DOCKER_IMAGE) { $env:DOCKER_IMAGE } else { "locally-api" }
$DockerTag = if ($env:DOCKER_TAG) { $env:DOCKER_TAG } else { "latest" }
$DockerFullName = "$DockerRegistry/$DockerNamespace/$DockerImage"

# Colors for output (PowerShell compatible)
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"
$White = "White"

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# Function to check if container exists
function Test-ContainerExists {
    param([string]$ContainerName)
    $container = docker ps -a --format "table {{.Names}}" | Select-String "^$ContainerName$"
    return $container -ne $null
}

# Function to check if container is running
function Test-ContainerRunning {
    param([string]$ContainerName)
    $container = docker ps --format "table {{.Names}}" | Select-String "^$ContainerName$"
    return $container -ne $null
}

# Function to get docker-compose version
function Get-DockerComposeVersion {
    try {
        $version = docker-compose --version 2>$null
        if ($version -match "docker-compose version (\d+\.\d+\.\d+)") {
            return $matches[1]
        }
        return $null
    }
    catch {
        return $null
    }
}

# Function to run container with direct Docker commands
function Start-ContainerDirect {
    Write-Status "Using direct Docker commands (bypassing docker-compose)"
    
    # Build the image
    Write-Status "Building Docker image..."
    docker build -f "$ApiDir\Dockerfile" -t locally-api:latest $ProjectRoot
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build Docker image"
        exit 1
    }
    
    # Create network if it doesn't exist
    $network = docker network ls --format "table {{.Name}}" | Select-String "^locally-network$"
    if ($network -eq $null) {
        Write-Status "Creating Docker network..."
        docker network create locally-network
    }
    
    # Create volume if it doesn't exist
    $volume = docker volume ls --format "table {{.Name}}" | Select-String "^locally-data$"
    if ($volume -eq $null) {
        Write-Status "Creating Docker volume..."
        docker volume create locally-data
    }
    
    # Run the container
    Write-Status "Starting container..."
    docker run -d --name locally-api `
        --network locally-network `
        -p 8080:8080 `
        -v locally-data:/app/data `
        -e LOCALLY_DATABASE_TYPE=sqlite `
        -e LOCALLY_DATABASE_STORAGE_PATH=/app/data/locally.db `
        -e LOCALLY_SERVER_BIND_TO=0.0.0.0 `
        -e LOCALLY_SERVER_API_PORT=8080 `
        -e LOCALLY_JWT_AUTH_SECRET=your-jwt-secret-here-change-in-production `
        -e LOCALLY_ENCRYPTION_MASTER_SECRET=your-encryption-master-secret-change-in-production `
        -e LOCALLY_ENCRYPTION_GLOBAL_SECRET=your-encryption-global-secret-change-in-production `
        -e LOCALLY_MESSAGE_PROCESSOR_POLL_INTERVAL=10s `
        -e LOCALLY_MESSAGE_PROCESSOR_PROCESSING_TIMEOUT=30m `
        -e LOCALLY_MESSAGE_PROCESSOR_DEFAULT_MAX_RETRIES=3 `
        -e LOCALLY_MESSAGE_PROCESSOR_RECOVERY_ENABLED=true `
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_ENABLED=true `
        -e LOCALLY_MESSAGE_PROCESSOR_MAX_PROCESSING_AGE=1h `
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_MAX_AGE=24h `
        -e LOCALLY_MESSAGE_PROCESSOR_CLEANUP_INTERVAL=1h `
        -e LOCALLY_MESSAGE_PROCESSOR_KEEP_COMPLETE_MESSAGES=true `
        -e LOCALLY_MESSAGE_PROCESSOR_DEBUG=false `
        ${DockerFullName}:${DockerTag}
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Container started successfully"
        Write-Status "API will be available at http://localhost:8080"
        Write-Status "Use 'docker logs locally-api' to view logs"
    } else {
        Write-Error "Failed to start container"
        exit 1
    }
}

# Function to stop container
function Stop-Container {
    if (Test-ContainerExists "locally-api") {
        Write-Status "Stopping container..."
        docker stop locally-api
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Container stopped"
        } else {
            Write-Error "Failed to stop container"
        }
    } else {
        Write-Warning "Container 'locally-api' does not exist"
    }
}

# Function to remove container
function Remove-Container {
    if (Test-ContainerExists "locally-api") {
        Write-Status "Removing container..."
        docker rm -f locally-api
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Container removed"
        } else {
            Write-Error "Failed to remove container"
        }
    } else {
        Write-Warning "Container 'locally-api' does not exist"
    }
}

# Function to show container logs
function Show-ContainerLogs {
    if (Test-ContainerExists "locally-api") {
        Write-Status "Showing container logs..."
        docker logs locally-api
    } else {
        Write-Warning "Container 'locally-api' does not exist"
    }
}

# Function to show container status
function Show-ContainerStatus {
    Write-Status "Container status:"
    if (Test-ContainerExists "locally-api") {
        if (Test-ContainerRunning "locally-api") {
            Write-Success "Container 'locally-api' is running"
        } else {
            Write-Warning "Container 'locally-api' exists but is not running"
        }
        docker ps -a --filter "name=locally-api" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    } else {
        Write-Warning "Container 'locally-api' does not exist"
    }
}

# Function to clean up everything
function Clean-Container {
    Write-Status "Cleaning up containers, networks, and volumes..."
    
    # Stop and remove container
    if (Test-ContainerExists "locally-api") {
        docker stop locally-api 2>$null
        docker rm -f locally-api 2>$null
    }
    
    # Remove network
    docker network rm locally-network 2>$null
    
    # Remove volume (WARNING: This will delete all data!)
    $response = Read-Host "Do you want to delete the database volume? This will delete all data! (y/N)"
    if ($response -eq "y" -or $response -eq "Y") {
        docker volume rm locally-data 2>$null
        Write-Success "Database volume removed"
    } else {
        Write-Status "Database volume preserved"
    }
    
    Write-Success "Cleanup completed"
}

# Main execution
Write-Status "Docker Run Script for Locally API (Windows PowerShell)"

switch ($Action) {
    "run" {
        # Check if container is already running
        if (Test-ContainerRunning "locally-api") {
            Write-Warning "Container 'locally-api' is already running"
            Write-Status "Use 'stop' to stop it first, or 'restart' to restart it"
            exit 0
        }
        
        # Check if container exists but is stopped
        if (Test-ContainerExists "locally-api") {
            Write-Status "Container exists but is stopped. Starting it..."
            docker start locally-api
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Container started successfully"
            } else {
                Write-Error "Failed to start existing container"
                exit 1
            }
        } else {
            # Check docker-compose version
            $composeVersion = Get-DockerComposeVersion
            if ($composeVersion -and $composeVersion -lt "2.0.0") {
                Write-Warning "Detected docker-compose version $composeVersion (known to have issues)"
                Write-Status "Using direct Docker commands instead"
                Start-ContainerDirect
            } else {
                Write-Status "Using docker-compose..."
                Push-Location $ApiDir
                try {
                    # Set build arguments for docker-compose
                    $env:VERSION = if ($env:VERSION) { $env:VERSION } else { "0.0.0" }
                    $env:BUILD_TIME = if ($env:BUILD_TIME) { $env:BUILD_TIME } else { (Get-Date -Format "yyyy-MM-dd_HH:mm:ss") }
                    $env:GIT_COMMIT = if ($env:GIT_COMMIT) { $env:GIT_COMMIT } else { 
                        try { git rev-parse --short HEAD 2>$null } catch { "unknown" }
                    }
                    docker-compose up -d --build
                    if ($LASTEXITCODE -eq 0) {
                        Write-Success "Container started successfully with docker-compose"
                    } else {
                        Write-Warning "docker-compose failed, trying direct Docker commands"
                        Start-ContainerDirect
                    }
                }
                finally {
                    Pop-Location
                }
            }
        }
    }
    
    "stop" {
        Stop-Container
    }
    
    "restart" {
        Stop-Container
        Start-Sleep -Seconds 2
        & $MyInvocation.MyCommand.Path "run"
    }
    
    "logs" {
        Show-ContainerLogs
    }
    
    "status" {
        Show-ContainerStatus
    }
    
    "clean" {
        Clean-Container
    }
    
    default {
        Write-Error "Unknown action: $Action"
        Write-Status "Available actions: run, stop, restart, logs, status, clean"
        exit 1
    }
} 