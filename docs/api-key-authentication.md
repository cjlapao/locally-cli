# API Key Authentication

This document describes the API key authentication system implemented in the locally-cli project. The system provides secure, persistent API key authentication alongside the existing username/password authentication.

## Overview

The API key system provides:
- **Secure Generation**: Cryptographically secure API keys using `crypto/rand`
- **Database Storage**: Encrypted storage with proper hashing
- **Key Management**: Create, list, revoke, and delete API keys
- **Usage Tracking**: Monitor API key usage and performance
- **Tenant Isolation**: API keys can be scoped to specific tenants
- **Permission System**: Fine-grained permissions for different operations
- **Audit Trail**: Complete audit trail for security compliance

## Security Features

### 1. Secure Key Generation
- Uses `crypto/rand` for cryptographically secure random generation
- 32-byte random data encoded as base64
- Prefix format: `sk-locally-` for easy identification
- Example: `sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz`

### 2. Secure Storage
- API keys are hashed using the same encryption service as passwords
- Only key prefixes are stored in plain text for identification
- Full keys are never stored in the database
- Keys are only shown once upon creation

### 3. Access Control
- API keys are tied to specific users
- Users can only access their own API keys
- Tenant-based isolation for multi-tenant environments
- Permission-based access control

## API Endpoints

### Authentication

#### Login with API Key
```http
POST /v1/auth/login/api-key
Content-Type: application/json

{
  "api_key": "sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
  "tenant_id": "optional-tenant-id"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": null,
  "expires_at": "2024-01-15T10:30:00Z"
}
```

### API Key Management

#### Create API Key
```http
POST /v1/auth/api-keys
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "My API Key",
  "permissions": {
    "read": true,
    "write": false,
    "delete": false,
    "admin": false,
    "scopes": ["users", "devices"]
  },
  "tenant_id": "optional-tenant-id",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

Response:
```json
{
  "id": "uuid-here",
  "name": "My API Key",
  "api_key": "sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
  "key_prefix": "sk-locally-abc123",
  "permissions": {
    "read": true,
    "write": false,
    "delete": false,
    "admin": false,
    "scopes": ["users", "devices"]
  },
  "tenant_id": "optional-tenant-id",
  "expires_at": "2024-12-31T23:59:59Z",
  "created_at": "2024-01-15T10:30:00Z",
  "created_by": "user-id"
}
```

#### List API Keys
```http
GET /v1/auth/api-keys
Authorization: Bearer <jwt-token>
```

Response:
```json
{
  "api_keys": [
    {
      "id": "uuid-here",
      "name": "My API Key",
      "key_prefix": "sk-locally-abc123",
      "permissions": {
        "read": true,
        "write": false,
        "delete": false,
        "admin": false,
        "scopes": ["users", "devices"]
      },
      "tenant_id": "optional-tenant-id",
      "expires_at": "2024-12-31T23:59:59Z",
      "last_used_at": "2024-01-15T10:30:00Z",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "created_by": "user-id",
      "revoked_at": null,
      "revoked_by": null
    }
  ],
  "total": 1
}
```

#### Get API Key Details
```http
GET /v1/auth/api-keys/{id}
Authorization: Bearer <jwt-token>
```

#### Revoke API Key
```http
POST /v1/auth/api-keys/{id}/revoke
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "reason": "Security concern"
}
```

#### Delete API Key
```http
DELETE /v1/auth/api-keys/{id}
Authorization: Bearer <jwt-token>
```

## Database Schema

### API Keys Table
```sql
CREATE TABLE api_keys (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(255) NOT NULL,
    permissions TEXT DEFAULT 'read',
    tenant_id VARCHAR(255),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(255),
    revoked_at TIMESTAMP,
    revoked_by VARCHAR(255),
    revocation_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

### API Key Usage Table
```sql
CREATE TABLE api_key_usage (
    id VARCHAR(255) PRIMARY KEY,
    api_key_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    ip_address VARCHAR(255),
    user_agent TEXT,
    endpoint VARCHAR(255),
    method VARCHAR(10),
    status_code INTEGER,
    response_time_ms BIGINT,
    tenant_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

## Usage Examples

### Using API Key in HTTP Client
```go
package main

import (
    "net/http"
    "fmt"
)

func main() {
    client := &http.Client{}
    
    req, err := http.NewRequest("GET", "http://localhost:5000/api/v1/users", nil)
    if err != nil {
        panic(err)
    }
    
    // Add API key to Authorization header
    req.Header.Set("Authorization", "Bearer sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz")
    
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    fmt.Printf("Status: %s\n", resp.Status)
}
```

### Using API Key in cURL
```bash
curl -X GET "http://localhost:5000/api/v1/users" \
  -H "Authorization: Bearer sk-locally-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz"
```

### Creating API Key Programmatically
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "fmt"
)

func main() {
    // First, login to get a JWT token
    loginData := map[string]string{
        "username": "your-username",
        "password": "your-password",
    }
    
    loginJSON, _ := json.Marshal(loginData)
    loginResp, _ := http.Post("http://localhost:5000/api/v1/auth/login", 
                             "application/json", bytes.NewBuffer(loginJSON))
    
    var loginResult map[string]interface{}
    json.NewDecoder(loginResp.Body).Decode(&loginResult)
    token := loginResult["token"].(string)
    
    // Create API key
    apiKeyData := map[string]interface{}{
        "name": "My Service API Key",
        "permissions": map[string]interface{}{
            "read": true,
            "write": true,
            "delete": false,
            "admin": false,
            "scopes": []string{"users", "devices"},
        },
        "expires_at": "2024-12-31T23:59:59Z",
    }
    
    apiKeyJSON, _ := json.Marshal(apiKeyData)
    req, _ := http.NewRequest("POST", "http://localhost:5000/api/v1/auth/api-keys", 
                             bytes.NewBuffer(apiKeyJSON))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    fmt.Printf("API Key: %s\n", result["api_key"])
}
```

## Security Best Practices

### 1. Key Management
- **Store securely**: Never store API keys in plain text or version control
- **Rotate regularly**: Set expiration dates and rotate keys periodically
- **Limit scope**: Use minimal required permissions
- **Monitor usage**: Regularly review API key usage logs

### 2. Key Storage
- **Environment variables**: Store in secure environment variables
- **Secret management**: Use tools like HashiCorp Vault or AWS Secrets Manager
- **Access control**: Limit access to API keys to authorized personnel only

### 3. Key Usage
- **HTTPS only**: Always use HTTPS when transmitting API keys
- **Header-based**: Use Authorization header, never URL parameters
- **Logging**: Avoid logging API keys in application logs

### 4. Monitoring
- **Usage tracking**: Monitor API key usage patterns
- **Anomaly detection**: Set up alerts for unusual usage patterns
- **Audit logs**: Maintain complete audit trails

## Migration from Password Authentication

The API key system is designed to work alongside existing password authentication. To migrate:

1. **Gradual migration**: Start with new integrations using API keys
2. **Dual support**: Maintain password authentication for existing users
3. **User education**: Educate users about API key benefits
4. **Monitoring**: Monitor usage patterns during migration

## Troubleshooting

### Common Issues

1. **Invalid API key format**
   - Ensure API key starts with `sk-locally-`
   - Check for extra spaces or characters

2. **API key not found**
   - Verify the API key exists in the database
   - Check if the key has been revoked or expired

3. **Permission denied**
   - Verify the API key has required permissions
   - Check tenant access if using multi-tenant setup

4. **Rate limiting**
   - Check if API key usage exceeds limits
   - Review usage patterns for potential abuse

### Debug Information

Enable debug logging to troubleshoot authentication issues:

```bash
export LOCALLY_DEBUG=true
export LOCALLY_LOG_LEVEL=debug
```

## Future Enhancements

Planned improvements to the API key system:

1. **Rate Limiting**: Per-key rate limiting
2. **Webhook Support**: Notifications for key events
3. **Key Rotation**: Automatic key rotation
4. **Advanced Permissions**: More granular permission system
5. **Usage Analytics**: Detailed usage analytics and reporting 