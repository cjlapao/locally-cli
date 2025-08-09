package types

type AppContextKey string

const (
	RequestIDKey     AppContextKey = "x-request-id"
	CorrelationIDKey AppContextKey = "x-correlation-id"
	UserIDKey        AppContextKey = "x-user-id"
	UsernameKey      AppContextKey = "x-username"
	TenantIDKey      AppContextKey = "x-tenant-id"
	UserIPKey        AppContextKey = "x-user-ip"
	UserAgentKey     AppContextKey = "x-user-agent"
	StartTimeKey     AppContextKey = "x-start-time"
	MetadataKey      AppContextKey = "x-metadata"
	SecurityLevelKey AppContextKey = "x-security-level"
)
