package config

const (
	UnknownTenantID                      = "00000000-0000-0000-0000-000000000000"
	UnknownUserID                        = "00000000-0000-0000-0000-000000000000"
	DefaultSuperUserUserID               = "11111111-1111-1111-1111-111111111111"
	GlobalTenantID                       = "11111111-1111-1111-1111-111111111111"
	GlobalRootCertificateID              = "11111111-1111-1111-1111-111111111111"
	GlobalTenantName                     = "Global Tenant"
	SuperUserRole                        = "su"
	RootCertificateSlug                  = "locally-root"
	IntermediateCertificateSlug          = "locally-ca"
	ApiKeyPrefix                         = "sk-locally-"
	PasswordAllowedSpecialChars          = "!@#$%.?"
	SystemStoragePath                    = ".locally"
	DefaultPageSizeInt                   = 20
	DefaultPageSize                      = "20"
	DefaultRetentionDays                 = 90
	ApiKeyAuthorizationHeader            = "X-API-KEY"
	TenantIDHeader                       = "X-Tenant-ID"
	CorrelationIDHeader                  = "X-Correlation-ID"
	RequestIDHeader                      = "X-Request-ID"
	SecurityLevelHeader                  = "X-Security-Level"
	UserAgentHeader                      = "X-User-Agent"
	UserIPHeader                         = "X-User-IP"
	UsernameHeader                       = "X-Username"
	StartTimeHeader                      = "X-Start-Time"
	DefaultRootCertificateCommonName     = "Locally"
	DefaultCertificateSubDomain          = "Locally Root CA"
	DefaultLocallyDomain                 = "locally.local"
	DefaultCertificateFQDN               = "*.locally.local"
	DefaultCertificateExpiresInYears     = 10
	DefaultCertificateKeySize            = 4096
	DefaultCertificateSignatureAlgorithm = "SHA512"
	DefaultCertificateCountry            = "UK"
	DefaultCertificateState              = "London"
	DefaultCertificateCity               = "London"
	DefaultCertificateOrganization       = "Locally"
	DefaultCertificateOrganizationalUnit = "Locally IT"
	DefaultCertificateAdminEmailAddress  = "admin@locally.local"
	// UUIDs
	RoleSuperUserID   = "11111111-1111-1111-1111-111111111111"
	RoleAdminUserID   = "22222222-2222-2222-2222-222222222222"
	RoleManagerUserID = "33333333-3333-3333-3333-333333333333"
	RoleUserID        = "44444444-4444-4444-4444-444444444444"
	RoleAuditorUserID = "55555555-5555-5555-5555-555555555555"
	RoleGuestUserID   = "66666666-6666-6666-6666-666666666666"
	RoleNoneUserID    = "77777777-7777-7777-7777-777777777777"
)

const (
	DefaultStoragePath = ".locally-cli"
	DefaultStorageFile = "locally.db"
)

const (
	DebugKey    = "debug"
	LogLevelKey = "log_level"

	// Server configuration keys
	ServerAPIPortKey     = "server.api_port"
	ServerBindAddressKey = "server.bind_to"
	ServerBaseURLKey     = "server.base_url"
	ServerAPIPrefixKey   = "server.api_prefix"
	AuthRootPasswordKey  = "auth.root_password"
	JwtAuthSecretKey     = "jwt.auth_secret"
	JwtIssuerKey         = "jwt.issuer"

	// Encryption configuration keys
	EncryptionMasterSecretKey = "encryption.master_secret"
	EncryptionGlobalSecretKey = "encryption.global_secret"

	// Root User configuration keys
	RootUserUsernameKey = "root_user.username"
	RootUserPasswordKey = "root_user.password"

	// API Key configuration keys
	APIKey = "api.key"

	// Security configuration keys
	SecurityPasswordMinLengthKey        = "security.password.min_length"
	SecurityPasswordRequireNumberKey    = "security.password.require_number"
	SecurityPasswordRequireSpecialKey   = "security.password.require_special"
	SecurityPasswordRequireUppercaseKey = "security.password.require_uppercase"

	// Domain configuration keys
	DomainNameKey              = "domain.name"
	DomainAdminEmailAddressKey = "domain.admin_email_address"

	// Certificate configuration keys
	CertificateExpiresInYearsKey     = "certificate.expires_in_years"
	CertificateKeySizeKey            = "certificate.key_size"
	CertificateSignatureAlgorithmKey = "certificate.signature_algorithm"
	CertificateCountryKey            = "certificate.country"
	CertificateStateKey              = "certificate.state"
	CertificateCityKey               = "certificate.city"
	CertificateOrganizationKey       = "certificate.organization"
	CertificateOrganizationalUnitKey = "certificate.organizational_unit"
	CertificateAdminEmailAddressKey  = "certificate.admin_email_address"

	// Seeding configuration keys
	SeedDemoDataKey = "seeding.demo_data"

	// Cors configuration keys
	CorsAllowOriginsKey  = "cors.allow_origins"
	CorsAllowMethodsKey  = "cors.allow_methods"
	CorsAllowHeadersKey  = "cors.allow_headers"
	CorsExposeHeadersKey = "cors.expose_headers"

	// Database configuration keys
	DatabaseTypeKey        = "database.type"
	DatabaseStoragePathKey = "database.storage_path"
	DatabaseHostKey        = "database.host"
	DatabasePortKey        = "database.port"
	DatabaseDatabaseKey    = "database.database"
	DatabaseUsernameKey    = "database.username"
	DatabasePasswordKey    = "database.password"
	DatabaseSSLModeKey     = "database.ssl_mode"
	DatabaseMigrateKey     = "database.migrate"

	// Pagination configuration keys
	PaginationDefaultPageSizeKey = "pagination.default_page_size"

	// Message Processor configuration keys
	MessageProcessorPollIntervalKey         = "message_processor.poll_interval"
	MessageProcessorProcessingTimeoutKey    = "message_processor.processing_timeout"
	MessageProcessorDefaultMaxRetriesKey    = "message_processor.default_max_retries"
	MessageProcessorRecoveryEnabledKey      = "message_processor.recovery_enabled"
	MessageProcessorMaxProcessingAgeKey     = "message_processor.max_processing_age"
	MessageProcessorCleanupEnabledKey       = "message_processor.cleanup_enabled"
	MessageProcessorCleanupMaxAgeKey        = "message_processor.cleanup_max_age"
	MessageProcessorCleanupIntervalKey      = "message_processor.cleanup_interval"
	MessageProcessorKeepCompleteMessagesKey = "message_processor.keep_complete_messages"
	MessageProcessorDebugKey                = "message_processor.debug"

	// Activity configuration keys
	ActivityRetentionDaysKey = "activity.retention_days"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	TenantIDContextKey ContextKey = "tenant_id"
)

// Environment variables
const (
	EnvPrefix      = "LOCALLY_"
	EnvProduction  = "production"
	EnvDevelopment = "development"
	DebugEnvKey    = EnvPrefix + "DEBUG"
	LogLevelEnvKey = EnvPrefix + "LOG_LEVEL"

	// Server configuration keys
	ServerAPIPortEnvKey          = EnvPrefix + "SERVER_API_PORT"
	ServerBindAddressEnvKey      = EnvPrefix + "SERVER_BIND_TO"
	ServerBaseURLEnvKey          = EnvPrefix + "SERVER_BASE_URL"
	ServerAPIPrefixEnvKey        = EnvPrefix + "SERVER_API_PREFIX"
	AuthRootPasswordEnvKey       = EnvPrefix + "AUTH_ROOT_PASSWORD"
	JwtAuthSecretEnvKey          = EnvPrefix + "JWT_AUTH_SECRET"
	JwtIssuerEnvKey              = EnvPrefix + "JWT_ISSUER"
	EncryptionMasterSecretEnvKey = EnvPrefix + "ENCRYPTION_MASTER_SECRET"
	EncryptionGlobalSecretEnvKey = EnvPrefix + "ENCRYPTION_GLOBAL_SECRET"

	// Root User configuration keys
	RootUserUsernameEnvKey = EnvPrefix + "ROOT_USER_USERNAME"
	RootUserPasswordEnvKey = EnvPrefix + "ROOT_USER_PASSWORD"
	SeedDemoDataEnvKey     = EnvPrefix + "SEED_DEMO_DATA"
	APIKeyEnvKey           = EnvPrefix + "API_KEY"

	// Domain configuration keys
	DomainNameEnvKey              = EnvPrefix + "DOMAIN_NAME"
	DomainAdminEmailAddressEnvKey = EnvPrefix + "DOMAIN_ADMIN_EMAIL_ADDRESS"

	// Security configuration keys
	SecurityPasswordMinLengthEnvKey        = EnvPrefix + "SECURITY_PASSWORD_MIN_LENGTH"
	SecurityPasswordRequireNumberEnvKey    = EnvPrefix + "SECURITY_PASSWORD_REQUIRE_NUMBER"
	SecurityPasswordRequireSpecialEnvKey   = EnvPrefix + "SECURITY_PASSWORD_REQUIRE_SPECIAL"
	SecurityPasswordRequireUppercaseEnvKey = EnvPrefix + "SECURITY_PASSWORD_REQUIRE_UPPERCASE"

	// Certificate configuration keys
	CertificateExpiresInYearsEnvKey     = EnvPrefix + "CERTIFICATE_EXPIRES_IN_YEARS"
	CertificateKeySizeEnvKey            = EnvPrefix + "CERTIFICATE_KEY_SIZE"
	CertificateSignatureAlgorithmEnvKey = EnvPrefix + "CERTIFICATE_SIGNATURE_ALGORITHM"
	CertificateCountryEnvKey            = EnvPrefix + "CERTIFICATE_COUNTRY"
	CertificateStateEnvKey              = EnvPrefix + "CERTIFICATE_STATE"
	CertificateCityEnvKey               = EnvPrefix + "CERTIFICATE_CITY"
	CertificateOrganizationEnvKey       = EnvPrefix + "CERTIFICATE_ORGANIZATION"
	CertificateOrganizationalUnitEnvKey = EnvPrefix + "CERTIFICATE_ORGANIZATIONAL_UNIT"
	CertificateAdminEmailAddressEnvKey  = EnvPrefix + "CERTIFICATE_ADMIN_EMAIL_ADDRESS"

	// Cors configuration keys
	CorsAllowOriginsEnvKey  = EnvPrefix + "CORS_ALLOW_ORIGINS"
	CorsAllowMethodsEnvKey  = EnvPrefix + "CORS_ALLOW_METHODS"
	CorsAllowHeadersEnvKey  = EnvPrefix + "CORS_ALLOW_HEADERS"
	CorsExposeHeadersEnvKey = EnvPrefix + "CORS_EXPOSE_HEADERS"

	// Database configuration keys
	DatabaseTypeEnvKey        = EnvPrefix + "DATABASE_TYPE"
	DatabaseStoragePathEnvKey = EnvPrefix + "DATABASE_STORAGE_PATH"
	DatabaseHostEnvKey        = EnvPrefix + "DATABASE_HOST"
	DatabasePortEnvKey        = EnvPrefix + "DATABASE_PORT"
	DatabaseDatabaseEnvKey    = EnvPrefix + "DATABASE_DATABASE"
	DatabaseUsernameEnvKey    = EnvPrefix + "DATABASE_USERNAME"
	DatabasePasswordEnvKey    = EnvPrefix + "DATABASE_PASSWORD"
	DatabaseSSLModeEnvKey     = EnvPrefix + "DATABASE_SSL_MODE"
	DatabaseMigrateEnvKey     = EnvPrefix + "DATABASE_MIGRATE"

	// Message Processor configuration keys
	MessageProcessorPollIntervalEnvKey         = EnvPrefix + "MESSAGE_PROCESSOR_POLL_INTERVAL"
	MessageProcessorProcessingTimeoutEnvKey    = EnvPrefix + "MESSAGE_PROCESSOR_PROCESSING_TIMEOUT"
	MessageProcessorDefaultMaxRetriesEnvKey    = EnvPrefix + "MESSAGE_PROCESSOR_DEFAULT_MAX_RETRIES"
	MessageProcessorRecoveryEnabledEnvKey      = EnvPrefix + "MESSAGE_PROCESSOR_RECOVERY_ENABLED"
	MessageProcessorMaxProcessingAgeEnvKey     = EnvPrefix + "MESSAGE_PROCESSOR_MAX_PROCESSING_AGE"
	MessageProcessorCleanupEnabledEnvKey       = EnvPrefix + "MESSAGE_PROCESSOR_CLEANUP_ENABLED"
	MessageProcessorCleanupMaxAgeEnvKey        = EnvPrefix + "MESSAGE_PROCESSOR_CLEANUP_MAX_AGE"
	MessageProcessorCleanupIntervalEnvKey      = EnvPrefix + "MESSAGE_PROCESSOR_CLEANUP_INTERVAL"
	MessageProcessorKeepCompleteMessagesEnvKey = EnvPrefix + "MESSAGE_PROCESSOR_KEEP_COMPLETE_MESSAGES"
	MessageProcessorDebugEnvKey                = EnvPrefix + "MESSAGE_PROCESSOR_DEBUG"

	// Pagination configuration keys
	PaginationDefaultPageSizeEnvKey = EnvPrefix + "PAGINATION_DEFAULT_PAGE_SIZE"

	// Activity configuration keys
	ActivityRetentionDaysEnvKey = EnvPrefix + "ACTIVITY_RETENTION_DAYS"
)

// Flags
const (
	FlagDebug                            = "debug"
	FlagLogLevel                         = "log-level"
	FlagAPIPort                          = "api-port"
	FlagBindTo                           = "bind-to"
	FlagBaseURL                          = "base-url"
	FlagAPIPrefix                        = "api-prefix"
	FlagRootPassword                     = "root-password"
	FlagJwtAuthSecret                    = "jwt-auth-secret"
	FlagJwtIssuer                        = "jwt-issuer"
	FlagEncryptionMasterSecret           = "encryption-master-secret"
	FlagEncryptionGlobalSecret           = "encryption-global-secret"
	FlagRootUserUsername                 = "root-user-username"
	FlagRootUserPassword                 = "root-user-password"
	FlagSecurityPasswordMinLength        = "security-password-min-length"
	FlagSecurityPasswordRequireNumber    = "security-password-require-number"
	FlagSecurityPasswordRequireSpecial   = "security-password-require-special"
	FlagSecurityPasswordRequireUppercase = "security-password-require-uppercase"
	FlagSeedDemoData                     = "seed-demo-data"
	FlagCorsAllowOrigins                 = "cors-allow-origins"
	FlagCorsAllowMethods                 = "cors-allow-methods"
	FlagCorsAllowHeaders                 = "cors-allow-headers"
	FlagCorsExposeHeaders                = "cors-expose-headers"
	FlagDatabaseType                     = "database-type"
	FlagDatabaseStoragePath              = "database-storage-path"
	FlagDatabaseHost                     = "database-host"
	FlagDatabasePort                     = "database-port"
	FlagDatabaseDatabase                 = "database-database"
	FlagDatabaseUsername                 = "database-username"
	FlagDatabasePassword                 = "database-password"
	FlagDatabaseSSLMode                  = "database-ssl-mode"
	FlagDatabaseMigrate                  = "database-migrate"

	// Message Processor flags
	FlagMessageProcessorPollInterval         = "message-processor-poll-interval"
	FlagMessageProcessorProcessingTimeout    = "message-processor-processing-timeout"
	FlagMessageProcessorDefaultMaxRetries    = "message-processor-default-max-retries"
	FlagMessageProcessorRecoveryEnabled      = "message-processor-recovery-enabled"
	FlagMessageProcessorMaxProcessingAge     = "message-processor-max-processing-age"
	FlagMessageProcessorCleanupEnabled       = "message-processor-cleanup-enabled"
	FlagMessageProcessorCleanupMaxAge        = "message-processor-cleanup-max-age"
	FlagMessageProcessorCleanupInterval      = "message-processor-cleanup-interval"
	FlagMessageProcessorKeepCompleteMessages = "message-processor-keep-complete-messages"
	FlagMessageProcessorDebug                = "message-processor-debug"

	// Pagination flags
	FlagPaginationDefaultPageSize = "pagination-default-page-size"

	// Activity flags
	FlagActivityRetentionDays = "activity-retention-days"
)
