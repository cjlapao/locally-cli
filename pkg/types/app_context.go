package types

type AppContextKey string

const (
	RequestIDKey     AppContextKey = "X-Request-ID"
	CorrelationIDKey AppContextKey = "X-Correlation-ID"
	UserIDKey        AppContextKey = "X-User-ID"
	UsernameKey      AppContextKey = "X-Username"
	TenantIDKey      AppContextKey = "X-Tenant-ID"
	UserIPKey        AppContextKey = "X-User-IP"
	UserAgentKey     AppContextKey = "X-User-Agent"
	StartTimeKey     AppContextKey = "X-Start-Time"
	MetadataKey      AppContextKey = "X-Metadata"
	SecurityLevelKey AppContextKey = "X-Security-Level"
	UserKey          AppContextKey = "X-User"
)
