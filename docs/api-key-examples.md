# API Key Authentication Examples

This document provides practical examples of how to use the API key authentication system in the locally-cli project.

## Creating an API Key

### Using the API

```bash
# First, authenticate with username/password to get a token
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password",
    "tenant_id": "your-tenant"
  }'

# Use the token to create an API key
curl -X POST http://localhost:8080/v1/auth/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "My API Key",
    "permissions": {
      "read": true,
      "write": true,
      "delete": false,
      "admin": false,
      "scopes": ["users", "devices"]
    },
    "tenant_id": "your-tenant",
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

### Response Example

```json
{
  "id": "api-key-uuid",
  "name": "My API Key",
  "api_key": "sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
  "key_prefix": "sk-locally-abc123de",
  "permissions": {
    "read": true,
    "write": true,
    "delete": false,
    "admin": false,
    "scopes": ["users", "devices"]
  },
  "tenant_id": "your-tenant",
  "expires_at": "2024-12-31T23:59:59Z",
  "created_at": "2024-01-15T10:30:00Z",
  "created_by": "your-username"
}
```

## Authenticating with an API Key

```bash
curl -X POST http://localhost:8080/v1/auth/login/api-key \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
    "tenant_id": "your-tenant"
  }'
```

### Response Example

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

## Listing API Keys

```bash
curl -X GET http://localhost:8080/v1/auth/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Response Example

```json
{
  "api_keys": [
    {
      "id": "api-key-uuid-1",
      "name": "My API Key",
      "key_prefix": "sk-locally-abc123de",
      "permissions": {
        "read": true,
        "write": true,
        "delete": false,
        "admin": false,
        "scopes": ["users", "devices"]
      },
      "tenant_id": "your-tenant",
      "expires_at": "2024-12-31T23:59:59Z",
      "last_used_at": "2024-01-15T14:30:00Z",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "created_by": "your-username",
      "revoked_at": null,
      "revoked_by": null
    }
  ],
  "total": 1
}
```

## Revoking an API Key

```bash
curl -X DELETE http://localhost:8080/v1/auth/api-keys/api-key-uuid \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Security concern"
  }'
```

## Using API Keys in Go Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/cjlapao/locally-cli/internal/auth"
    "github.com/cjlapao/locally-cli/internal/database/stores"
)

func main() {
    // Initialize the auth service
    authStore := stores.GetAuthDataStoreInstance()
    authService := auth.NewService(auth.AuthServiceConfig{
        SecretKey: "your-secret-key",
    }, authStore)
    
    // Create an API key
    ctx := context.Background()
    userID := "user-uuid"
    
    req := auth.CreateAPIKeyRequest{
        Name: "My Service API Key",
        Permissions: types.APIKeyPermissions{
            Read:   true,
            Write:  true,
            Delete: false,
            Admin:  false,
            Scopes: []string{"users"},
        },
        TenantID: "my-tenant",
    }
    
    response, err := authService.CreateAPIKey(ctx, userID, req, userID)
    if err != nil {
        log.Fatalf("Failed to create API key: %v", err)
    }
    
    fmt.Printf("Created API key: %s\n", response.APIKey)
    fmt.Printf("Key prefix: %s\n", response.KeyPrefix)
    
    // Authenticate with the API key
    creds := auth.APIKeyCredentials{
        APIKey:   response.APIKey,
        TenantID: "my-tenant",
    }
    
    token, err := authService.AuthenticateWithAPIKey(creds)
    if err != nil {
        log.Fatalf("Failed to authenticate: %v", err)
    }
    
    fmt.Printf("Authentication successful! Token: %s\n", token.Token)
}
```

## Security Best Practices

1. **Store API Keys Securely**: Never store API keys in plain text or commit them to version control
2. **Use Environment Variables**: Store API keys in environment variables
3. **Rotate Keys Regularly**: Set expiration dates and rotate keys periodically
4. **Limit Permissions**: Only grant the minimum permissions necessary
5. **Monitor Usage**: Regularly check API key usage logs
6. **Revoke Compromised Keys**: Immediately revoke any keys that may have been compromised

## Environment Variable Example

```bash
# Set your API key as an environment variable
export LOCALLY_API_KEY="sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz"
export LOCALLY_TENANT_ID="your-tenant"

# Use in your application
curl -X POST http://localhost:8080/v1/auth/login/api-key \
  -H "Content-Type: application/json" \
  -d "{
    \"api_key\": \"$LOCALLY_API_KEY\",
    \"tenant_id\": \"$LOCALLY_TENANT_ID\"
  }"
```

## Error Handling

Common error responses:

```json
{
  "error": "invalid API key format",
  "message": "API key must start with 'sk-locally-'"
}
```

```json
{
  "error": "API key is revoked",
  "message": "This API key has been revoked and cannot be used"
}
```

```json
{
  "error": "API key has expired",
  "message": "This API key has expired and cannot be used"
}
```

```json
{
  "error": "API key not valid for this tenant",
  "message": "This API key is not authorized for the specified tenant"
}
``` 