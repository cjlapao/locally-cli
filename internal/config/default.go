package config

func DefaultConfig() *Config {
	return &Config{
		Items: []ConfigItem{
			{Key: DebugKey, Value: "false", EnvName: DebugEnvKey, FlagName: FlagDebug},
			{Key: LogLevelKey, Value: "info", EnvName: LogLevelEnvKey, FlagName: FlagLogLevel},

			{Key: ServerAPIPortKey, Value: "5000", EnvName: ServerAPIPortEnvKey, FlagName: FlagAPIPort},
			{Key: ServerBindAddressKey, Value: "0.0.0.0", EnvName: ServerBindAddressEnvKey, FlagName: FlagBindTo},
			{Key: ServerBaseURLKey, Value: "http://localhost:5000", EnvName: ServerBaseURLEnvKey, FlagName: FlagBaseURL},
			{Key: ServerAPIPrefixKey, Value: "/api", EnvName: ServerAPIPrefixEnvKey, FlagName: FlagAPIPrefix},
			{Key: AuthRootPasswordKey, Value: "root", EnvName: AuthRootPasswordEnvKey, FlagName: FlagRootPassword},
			{Key: JwtAuthSecretKey, Value: "secret", EnvName: JwtAuthSecretEnvKey, FlagName: FlagJwtAuthSecret},
			{Key: JwtIssuerKey, Value: "locally-cli", EnvName: JwtIssuerEnvKey, FlagName: FlagJwtIssuer},
			{Key: EncryptionMasterSecretKey, Value: "default-master-secret-change-in-production", EnvName: EncryptionMasterSecretEnvKey, FlagName: FlagEncryptionMasterSecret},
			{Key: EncryptionGlobalSecretKey, Value: "default-global-secret-change-in-production", EnvName: EncryptionGlobalSecretEnvKey, FlagName: FlagEncryptionGlobalSecret},

			// Root User Default Values
			{Key: RootUserUsernameKey, Value: "root", EnvName: RootUserUsernameEnvKey, FlagName: FlagRootUserUsername},
			{Key: RootUserPasswordKey, Value: "root", EnvName: RootUserPasswordEnvKey, FlagName: FlagRootUserPassword},

			// Seeding Default Values
			{Key: SeedDemoDataKey, Value: "false", EnvName: SeedDemoDataEnvKey, FlagName: FlagSeedDemoData},

			// Pagination Default Values
			{Key: PaginationDefaultPageSizeKey, Value: "20", EnvName: PaginationDefaultPageSizeEnvKey, FlagName: FlagPaginationDefaultPageSize},

			// Security Default Values
			{Key: SecurityPasswordMinLengthKey, Value: "8", EnvName: SecurityPasswordMinLengthEnvKey, FlagName: FlagSecurityPasswordMinLength},
			{Key: SecurityPasswordRequireNumberKey, Value: "true", EnvName: SecurityPasswordRequireNumberEnvKey, FlagName: FlagSecurityPasswordRequireNumber},
			{Key: SecurityPasswordRequireSpecialKey, Value: "true", EnvName: SecurityPasswordRequireSpecialEnvKey, FlagName: FlagSecurityPasswordRequireSpecial},
			{Key: SecurityPasswordRequireUppercaseKey, Value: "true", EnvName: SecurityPasswordRequireUppercaseEnvKey, FlagName: FlagSecurityPasswordRequireUppercase},

			// API Key
			{Key: APIKey, Value: "sk-locally-"},

			// Cors
			{Key: CorsAllowOriginsKey, Value: "http://localhost:3000, http://127.0.0.1:3000", EnvName: CorsAllowOriginsEnvKey, FlagName: FlagCorsAllowOrigins},
			{Key: CorsAllowMethodsKey, Value: "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD", EnvName: CorsAllowMethodsEnvKey, FlagName: FlagCorsAllowMethods},
			{Key: CorsAllowHeadersKey, Value: "Accept, Accept-Language, Content-Type, Content-Language, Origin, Authorization, X-Requested-With, X-Request-ID, X-HTTP-Method-Override, Cache-Control, X-Tenant-ID", EnvName: CorsAllowHeadersEnvKey, FlagName: FlagCorsAllowHeaders},
			{Key: CorsExposeHeadersKey, Value: "X-Request-ID", EnvName: CorsExposeHeadersEnvKey, FlagName: FlagCorsExposeHeaders},

			// Database Default Values
			{Key: DatabaseTypeKey, Value: "sqlite", EnvName: DatabaseTypeEnvKey, FlagName: FlagDatabaseType},
			{Key: DatabaseStoragePathKey, Value: "", EnvName: DatabaseStoragePathEnvKey, FlagName: FlagDatabaseStoragePath},
			{Key: DatabaseHostKey, Value: "localhost", EnvName: DatabaseHostEnvKey, FlagName: FlagDatabaseHost},
			{Key: DatabasePortKey, Value: "5432", EnvName: DatabasePortEnvKey, FlagName: FlagDatabasePort},
			{Key: DatabaseDatabaseKey, Value: "locally", EnvName: DatabaseDatabaseEnvKey, FlagName: FlagDatabaseDatabase},
			{Key: DatabaseUsernameKey, Value: "locally", EnvName: DatabaseUsernameEnvKey, FlagName: FlagDatabaseUsername},
			{Key: DatabasePasswordKey, Value: "locally", EnvName: DatabasePasswordEnvKey, FlagName: FlagDatabasePassword},
			{Key: DatabaseSSLModeKey, Value: "false", EnvName: DatabaseSSLModeEnvKey, FlagName: FlagDatabaseSSLMode},
			{Key: DatabaseMigrateKey, Value: "false", EnvName: DatabaseMigrateEnvKey, FlagName: FlagDatabaseMigrate},

			{Key: MessageProcessorDefaultMaxRetriesKey, Value: "3", EnvName: MessageProcessorDefaultMaxRetriesEnvKey, FlagName: FlagMessageProcessorDefaultMaxRetries},
			{Key: MessageProcessorPollIntervalKey, Value: "10s", EnvName: MessageProcessorPollIntervalEnvKey, FlagName: FlagMessageProcessorPollInterval},
			{Key: MessageProcessorProcessingTimeoutKey, Value: "30m", EnvName: MessageProcessorProcessingTimeoutEnvKey, FlagName: FlagMessageProcessorProcessingTimeout},
			{Key: MessageProcessorRecoveryEnabledKey, Value: "true", EnvName: MessageProcessorRecoveryEnabledEnvKey, FlagName: FlagMessageProcessorRecoveryEnabled},
			{Key: MessageProcessorMaxProcessingAgeKey, Value: "1h", EnvName: MessageProcessorMaxProcessingAgeEnvKey, FlagName: FlagMessageProcessorMaxProcessingAge},
			{Key: MessageProcessorCleanupEnabledKey, Value: "true", EnvName: MessageProcessorCleanupEnabledEnvKey, FlagName: FlagMessageProcessorCleanupEnabled},
			{Key: MessageProcessorCleanupMaxAgeKey, Value: "24h", EnvName: MessageProcessorCleanupMaxAgeEnvKey, FlagName: FlagMessageProcessorCleanupMaxAge},
			{Key: MessageProcessorCleanupIntervalKey, Value: "4h", EnvName: MessageProcessorCleanupIntervalEnvKey, FlagName: FlagMessageProcessorCleanupInterval},
			{Key: MessageProcessorKeepCompleteMessagesKey, Value: "false", EnvName: MessageProcessorKeepCompleteMessagesEnvKey, FlagName: FlagMessageProcessorKeepCompleteMessages},
			{Key: MessageProcessorDebugKey, Value: "false", EnvName: MessageProcessorDebugEnvKey, FlagName: FlagMessageProcessorDebug},
		},
	}
}
