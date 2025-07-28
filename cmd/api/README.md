# Locally API Docker Setup

This directory contains the Docker configuration for the Locally API service.

## Quick Start

### Using the Docker Run Script (Recommended)

The `docker-run.sh` script provides a robust way to manage the API container, handling docker-compose issues and providing better control:

```bash
# Start the API
./docker-run.sh run

# Stop the API
./docker-run.sh stop

# View logs
./docker-run.sh logs

# Check status
./docker-run.sh status

# Restart the API
./docker-run.sh restart

# Clean up everything (containers, volumes, networks)
./docker-run.sh clean
```

### Using Docker Compose

1. **Copy the environment template:**
   ```bash
   cp env.template .env
   ```

2. **Edit the environment file:**
   ```bash
   # Edit .env file with your specific values
   nano .env
   ```

3. **Build and run the service:**
   ```bash
   docker-compose up --build
   ```

4. **Run in background:**
   ```bash
   docker-compose up -d --build
   ```

5. **Stop the service:**
   ```bash
   docker-compose down
   ```

### Using Docker directly

1. **Copy and configure environment:**
   ```bash
   cp env.template .env
   # Edit .env with your values
   ```

2. **Build the image:**
   ```bash
   docker build -t locally-api -f Dockerfile ../..
   ```

3. **Run the container:**
   ```bash
   docker run -d \
     --name locally-api \
     -p 8080:8080 \
     --env-file .env \
     -v locally-data:/app/data \
     locally-api
   ```

## Docker Run Script Features

The `docker-run.sh` script provides several advantages:

- **Automatic docker-compose version detection** - Handles known issues with older docker-compose versions
- **Fallback to direct Docker commands** - Ensures compatibility across different environments
- **Robust container management** - Proper cleanup and status checking
- **Database persistence** - Preserves SQLite database across restarts
- **Cross-platform compatibility** - Works on Linux, macOS, and Windows

### Script Commands

| Command | Description |
|---------|-------------|
| `run` | Start the API container (default) |
| `stop` | Stop the running container |
| `clean` | Remove container, volumes, and networks |
| `logs` | Show container logs |
| `status` | Show container status |
| `restart` | Restart the container |

### Example Usage

```bash
# First time setup
./docker-run.sh run

# Check if it's running
./docker-run.sh status

# View logs if there are issues
./docker-run.sh logs

# Restart after configuration changes
./docker-run.sh restart

# Clean up for fresh start
./docker-run.sh clean
./docker-run.sh run
```

## Environment Configuration

### Environment Template

The `env.template` file contains all available environment variables with descriptions and default values. Copy this file to `.env` and modify the values as needed:

```bash
cp env.template .env
```

### Key Environment Variables

| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| **Server** | `LOCALLY_SERVER_API_PORT` | API server port | `8080` |
| **Server** | `LOCALLY_SERVER_BIND_TO` | Server bind address | `0.0.0.0` |
| **Security** | `LOCALLY_JWT_AUTH_SECRET` | JWT signing secret | **(required)** |
| **Security** | `LOCALLY_ENCRYPTION_MASTER_SECRET` | Encryption master secret | **(required)** |
| **Database** | `LOCALLY_DATABASE_TYPE` | Database type (sqlite/postgres) | `sqlite` |
| **Database** | `LOCALLY_DATABASE_STORAGE_PATH` | SQLite database path | `/app/data/locally.db` |
| **CORS** | `LOCALLY_CORS_ALLOW_ORIGINS` | Allowed origins | `*` |
| **Logging** | `LOCALLY_LOG_LEVEL` | Logging level | `info` |

### Required Environment Variables

These variables **must** be set in production:

- `LOCALLY_JWT_AUTH_SECRET` - JWT signing secret
- `LOCALLY_ENCRYPTION_MASTER_SECRET` - Encryption master secret  
- `LOCALLY_ENCRYPTION_GLOBAL_SECRET` - Encryption global secret
- `LOCALLY_AUTH_ROOT_PASSWORD` - Root user password

### Production Security Checklist

Before deploying to production, ensure you:

- [ ] Change all default secrets
- [ ] Use strong, unique passwords
- [ ] Configure proper CORS origins
- [ ] Enable SSL/TLS
- [ ] Use managed database service
- [ ] Set up proper logging
- [ ] Configure backups

## Configuration

### Environment Variables

The following environment variables can be configured:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOCALLY_DATABASE_TYPE` | Database type (sqlite/postgres) | sqlite |
| `LOCALLY_DATABASE_STORAGE_PATH` | SQLite database path | /app/data/locally.db |
| `LOCALLY_JWT_AUTH_SECRET` | JWT signing secret | (required) |
| `LOCALLY_SECURITY_PASSWORD_MIN_LENGTH` | Minimum password length | 8 |
| `LOCALLY_SECURITY_PASSWORD_REQUIRE_NUMBER` | Require numbers in password | true |
| `LOCALLY_SECURITY_PASSWORD_REQUIRE_SPECIAL` | Require special characters | true |
| `LOCALLY_SECURITY_PASSWORD_REQUIRE_UPPERCASE` | Require uppercase letters | true |
| `LOCALLY_SERVER_BIND_TO` | API host binding | 0.0.0.0 |
| `LOCALLY_SERVER_API_PORT` | API port | 8080 |
| `LOCALLY_LOG_LEVEL` | Logging level | info |

### Volumes

- `locally-data`: Persistent storage for the SQLite database

## Development

### Building for development

```bash
# Build with debug information
docker build --target builder -t locally-api-dev -f Dockerfile ../..

# Run with source code mounted for hot reloading
docker run -it --rm \
  -p 8080:8080 \
  -v $(pwd)/../../:/app \
  -w /app \
  locally-api-dev \
  go run ./cmd/api
```

### Debugging

```bash
# View logs using the script
./docker-run.sh logs

# Or using docker-compose
docker-compose logs -f locally-api

# Access container shell
docker-compose exec locally-api sh

# Check health status
docker-compose ps

# View environment variables
docker-compose exec locally-api env | grep LOCALLY
```

## Production Considerations

1. **Security:**
   - Change the JWT secret in production
   - Use environment-specific configuration
   - Consider using secrets management
   - Enable SSL/TLS

2. **Database:**
   - For production, consider using PostgreSQL instead of SQLite
   - Ensure proper backup strategies
   - Use managed database services when possible
   - Configure connection pooling

3. **Monitoring:**
   - The container includes health checks
   - Consider adding application metrics
   - Set up log aggregation
   - Monitor resource usage

4. **Scaling:**
   - The service is stateless and can be scaled horizontally
   - Use a load balancer for multiple instances
   - Consider using container orchestration (Kubernetes, Docker Swarm)

## Troubleshooting

### Common Issues

1. **Docker Compose ContainerConfig Error:**
   ```bash
   # This is a known issue with docker-compose 1.29.2
   # Use the docker-run.sh script instead
   ./docker-run.sh run
   ```

2. **Port already in use:**
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   # Or change the port in docker-compose.yml
   ```

3. **Permission issues:**
   ```bash
   # Fix volume permissions
   sudo chown -R $USER:$USER ./data
   ```

4. **Database connection issues:**
   ```bash
   # Check database file
   docker exec locally-api ls -la /app/data/
   ```

5. **Environment variable issues:**
   ```bash
   # Check environment variables
   docker exec locally-api env | grep LOCALLY
   ```

### Logs

```bash
# View application logs using the script
./docker-run.sh logs

# Or using docker-compose
docker-compose logs locally-api

# Follow logs in real-time
docker-compose logs -f locally-api

# View logs with timestamps
docker-compose logs -t locally-api
```

## API Endpoints

Once running, the API will be available at:
- **Health Check:** `http://localhost:8080/health`
- **API Documentation:** `http://localhost:8080/docs` (if available)
- **Version Info:** `http://localhost:8080/version`

## Environment Template Reference

The `env.template` file contains comprehensive documentation for all available environment variables, including:

- **Debug & Logging** - Debug mode and log levels
- **Server Configuration** - Port, bind address, base URL
- **Authentication & Security** - JWT secrets, encryption keys
- **Password Security Policy** - Password complexity requirements
- **CORS Configuration** - Cross-origin resource sharing settings
- **Database Configuration** - Database type and connection settings
- **Pagination** - Default page sizes
- **Message Processor** - Background job processing settings
- **Seeding & Demo Data** - Initial data population

Copy `env.template` to `.env` and customize the values for your environment. 