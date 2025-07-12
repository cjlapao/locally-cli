package types

const (
	RequestIDHeader = "X-Request-ID"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "x-request-id"
	UserIDKey    ContextKey = "x-user-id"
	TenantIDKey  ContextKey = "x-tenant-id"
	StartTimeKey ContextKey = "x-start-time"
	MetadataKey  ContextKey = "x-metadata"
)
