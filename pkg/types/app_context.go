package types

type AppContextKey string

const (
	RequestIDKey AppContextKey = "x-request-id"
	UserIDKey    AppContextKey = "x-user-id"
	UsernameKey  AppContextKey = "x-username"
	TenantIDKey  AppContextKey = "x-tenant-id"
	StartTimeKey AppContextKey = "x-start-time"
	MetadataKey  AppContextKey = "x-metadata"
)
