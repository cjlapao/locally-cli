# Swagger API Documentation

This project includes automatic Swagger/OpenAPI documentation generation for the API endpoints.

## Overview

The API uses [swaggo/swag](https://github.com/swaggo/swag) to automatically generate OpenAPI documentation from Go code comments. The documentation is served through the Swagger UI interface.

## Features

- **Automatic Documentation**: Documentation is generated from Go code comments
- **Interactive UI**: Swagger UI provides an interactive interface for testing API endpoints
- **Security Support**: Documents authentication methods (Bearer token, API key)
- **Request/Response Examples**: Shows expected request and response formats
- **Tagged Endpoints**: Endpoints are organized by functional areas

## Usage

### Generate Documentation

```bash
# Generate Swagger documentation
make swagger

# Or generate all documentation
make docs
```

### View Documentation

1. Start the API server:
   ```bash
   make dev
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger
   ```

### Documentation Structure

The API documentation is organized into the following tags:

- **Authentication**: Login, logout, token refresh, API key management
- **Users**: User management operations
- **Tenants**: Tenant management operations
- **Contexts**: Context and environment management
- **Messages**: Message processing and worker management
- **Certificates**: Certificate management operations
- **Environment**: Environment variable management
- **Events**: Event management and notifications
- **Workers**: Worker and task management
- **Health**: Health check and status endpoints

## Adding Documentation to New Endpoints

To add Swagger documentation to a new endpoint, add comments above the handler function:

```go
// @Summary      Brief description
// @Description  Detailed description
// @Tags         TagName
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Parameter description"
// @Param        request body RequestType true "Request body description"
// @Success      200  {object}  ResponseType
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /path/to/endpoint [method]
func (h *Handler) HandleEndpoint(w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}
```

### Available Annotations

- `@Summary`: Brief description of the endpoint
- `@Description`: Detailed description
- `@Tags`: Category for grouping endpoints
- `@Accept`: Expected content type for requests
- `@Produce`: Content type of responses
- `@Security`: Authentication method required
- `@Param`: Request parameters (path, query, body)
- `@Success`: Successful response format
- `@Failure`: Error response formats
- `@Router`: URL path and HTTP method

### Security

The API supports two authentication methods:

1. **Bearer Token**: JWT token in Authorization header
   ```
   Authorization: Bearer <token>
   ```

2. **API Key**: API key in X-API-Key header
   ```
   X-API-Key: <api-key>
   ```

## Configuration

Swagger configuration can be customized in the API server:

```go
swaggerConfig := api.SwaggerConfig{
    Enabled: true,
    Path:    "/swagger",
    Host:    "localhost:8080",
}
```

## Development Workflow

1. **Add Documentation**: Add Swagger comments to new handlers
2. **Generate Docs**: Run `make swagger` to regenerate documentation
3. **Test**: Start the server and verify documentation at `/swagger`
4. **Commit**: Include documentation updates in your commits

## Troubleshooting

### Documentation Not Updating

If changes don't appear in the Swagger UI:

1. Regenerate documentation: `make swagger`
2. Restart the API server: `make dev`
3. Clear browser cache and refresh the page

### Missing Endpoints

If endpoints don't appear in the documentation:

1. Check that Swagger comments are properly formatted
2. Verify the `@Router` path matches the actual route
3. Ensure the handler function is exported (capitalized)
4. Regenerate documentation: `make swagger`

### Build Errors

If you encounter build errors related to Swagger:

1. Install swag: `go install github.com/swaggo/swag/cmd/swag@latest`
2. Check that all imports are correct
3. Verify Swagger comments syntax

## Integration with CI/CD

The Swagger documentation can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Generate API Documentation
  run: make swagger
```

This ensures documentation is always up-to-date with the codebase. 