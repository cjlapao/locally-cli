# Locally CLI Scripts

This directory contains cross-platform scripts for managing the Locally CLI project.

## Available Scripts

### Docker Management Scripts

#### `docker-run.sh` (Linux/macOS)
Bash script for managing Docker containers on Linux and macOS systems.

**Usage:**
```bash
./scripts/docker-run.sh [action]
```

**Actions:**
- `run` - Start the API container (default)
- `stop` - Stop the running container
- `restart` - Stop and restart the container
- `logs` - Show container logs
- `status` - Show container status
- `clean` - Remove container, volumes, and networks

#### `docker-run.ps1` (Windows)
PowerShell script for managing Docker containers on Windows systems.

**Usage:**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/docker-run.ps1 [action]
```

**Actions:**
- `run` - Start the API container (default)
- `stop` - Stop the running container
- `restart` - Stop and restart the container
- `logs` - Show container logs
- `status` - Show container status
- `clean` - Remove container, volumes, and networks

## Features

### Cross-Platform Compatibility
- **Automatic platform detection** via Makefile
- **Platform-specific implementations** for optimal performance
- **Consistent interface** across all platforms

### Docker Management
- **Automatic docker-compose version detection** - Handles known issues with older docker-compose versions
- **Fallback to direct Docker commands** - Ensures compatibility across different environments
- **Robust container management** - Proper cleanup and status checking
- **Database persistence** - Preserves SQLite database across restarts

### Error Handling
- **Comprehensive error checking** - Validates Docker commands and provides clear error messages
- **Graceful fallbacks** - Falls back to direct Docker commands when docker-compose fails
- **User-friendly output** - Colored output with clear status messages

## Integration with Makefile

The Makefile automatically detects the platform and calls the appropriate script:

```bash
# These commands work on all platforms
make docker-run      # Start the API
make docker-stop     # Stop the API
make docker-logs     # View logs
make docker-status   # Check status
make docker-clean    # Clean up everything
```

## Platform-Specific Details

### Linux/macOS
- Uses bash scripting with Unix commands
- Compatible with most Linux distributions and macOS
- Uses `grep`, `cut`, and other Unix utilities

### Windows
- Uses PowerShell with Windows-specific commands
- Requires PowerShell execution policy to be set (handled by Makefile)
- Uses Windows path separators and PowerShell syntax

## Environment Variables

The scripts automatically set the following environment variables for the Docker container:

### Database Configuration
- `LOCALLY_DATABASE_TYPE=sqlite`
- `LOCALLY_DATABASE_STORAGE_PATH=/app/data/locally.db`

### Server Configuration
- `LOCALLY_SERVER_BIND_TO=0.0.0.0`
- `LOCALLY_SERVER_API_PORT=8080`

### Security (Change in Production!)
- `LOCALLY_JWT_AUTH_SECRET=your-jwt-secret-here-change-in-production`
- `LOCALLY_ENCRYPTION_MASTER_SECRET=your-encryption-master-secret-change-in-production`
- `LOCALLY_ENCRYPTION_GLOBAL_SECRET=your-encryption-global-secret-change-in-production`

### Message Processor Configuration
- `LOCALLY_MESSAGE_PROCESSOR_POLL_INTERVAL=10s`
- `LOCALLY_MESSAGE_PROCESSOR_PROCESSING_TIMEOUT=30m`
- `LOCALLY_MESSAGE_PROCESSOR_DEFAULT_MAX_RETRIES=3`
- `LOCALLY_MESSAGE_PROCESSOR_RECOVERY_ENABLED=true`
- `LOCALLY_MESSAGE_PROCESSOR_CLEANUP_ENABLED=true`
- `LOCALLY_MESSAGE_PROCESSOR_MAX_PROCESSING_AGE=1h`
- `LOCALLY_MESSAGE_PROCESSOR_CLEANUP_MAX_AGE=24h`
- `LOCALLY_MESSAGE_PROCESSOR_CLEANUP_INTERVAL=1h`
- `LOCALLY_MESSAGE_PROCESSOR_KEEP_COMPLETE_MESSAGES=true`
- `LOCALLY_MESSAGE_PROCESSOR_DEBUG=false`

## Troubleshooting

### Common Issues

1. **Docker not running**
   - Ensure Docker Desktop is running
   - Check with `docker --version`

2. **Permission denied (Linux/macOS)**
   - Make script executable: `chmod +x scripts/docker-run.sh`

3. **PowerShell execution policy (Windows)**
   - The Makefile handles this automatically
   - Manual fix: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

4. **Container already exists**
   - Use `clean` action to remove existing containers
   - Or use `restart` to stop and start

5. **Port already in use**
   - Check if another service is using port 8080
   - Stop the conflicting service or change the port in the script

### Debugging

1. **Check container status:**
   ```bash
   make docker-status
   ```

2. **View container logs:**
   ```bash
   make docker-logs
   ```

3. **Manual Docker commands:**
   ```bash
   docker ps -a
   docker logs locally-api
   ```

## Future Scripts

This directory is designed to accommodate additional scripts as the project grows. Planned additions may include:

- Database migration scripts
- Development environment setup scripts
- Testing automation scripts
- Deployment scripts
- Monitoring and health check scripts

## Contributing

When adding new scripts:

1. **Follow the naming convention**: `script-name.sh` for Linux/macOS, `script-name.ps1` for Windows
2. **Include comprehensive error handling**
3. **Add colored output** for better user experience
4. **Update this README** with usage instructions
5. **Test on all target platforms**
6. **Update the Makefile** if needed for integration 