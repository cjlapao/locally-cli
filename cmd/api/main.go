package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cjlapao/common-go/version"
	"github.com/cjlapao/locally-cli/internal/activity"
	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/api"
	api_handlers "github.com/cjlapao/locally-cli/internal/api/handlers"
	"github.com/cjlapao/locally-cli/internal/api_keys"
	api_keys_interfaces "github.com/cjlapao/locally-cli/internal/api_keys/interfaces"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/auth/handlers"
	auth_handlers "github.com/cjlapao/locally-cli/internal/auth/handlers"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	"github.com/cjlapao/locally-cli/internal/certificates"
	certificates_interfaces "github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/claim"
	claim_interfaces "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/seeds"
	"github.com/cjlapao/locally-cli/internal/database/seeds/migrations"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/encryption"
	"github.com/cjlapao/locally-cli/internal/environment"
	"github.com/cjlapao/locally-cli/internal/events"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/internal/role"
	rolesvc "github.com/cjlapao/locally-cli/internal/role/interfaces"
	"github.com/cjlapao/locally-cli/internal/system"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/tenant"
	tenant_interfaces "github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	"github.com/cjlapao/locally-cli/internal/user"
	user_interfaces "github.com/cjlapao/locally-cli/internal/user/interfaces"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/internal/vaults/configvault"
	"github.com/cjlapao/locally-cli/internal/workers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/interfaces"
)

// @title           Locally API
// @version         1.0
// @description     A comprehensive API for managing local development environments, contexts, and services.
// @termsOfService  https://locally.cloud/terms

// @contact.name   API Support
// @contact.url    https://locally.cloud/support
// @contact.email  support@locally.cloud

// @license.name  Fair Source License
// @license.url   https://locally.cloud/license

// @host      localhost:8080
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication

// @tag.name Authentication
// @tag.description Authentication and authorization endpoints

// @tag.name Users
// @tag.description User management operations

// @tag.name Tenants
// @tag.description Tenant management operations

// @tag.name Contexts
// @tag.description Context and environment management

// @tag.name Messages
// @tag.description Message processing and worker management

// @tag.name Certificates
// @tag.description Certificate management operations

// @tag.name Environment
// @tag.description Environment variable management

// @tag.name Events
// @tag.description Event management and notifications

// @tag.name Workers
// @tag.description Worker and task management

// @tag.name Health
// @tag.description Health check and status endpoints

var (
	versionSvc = version.Get()
	appVersion = "0.0.0" // This will be overridden by build flags
)

// Build-time variables (set by build flags)
var (
	Version   = "0.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	logging.Info("Starting locally API")
	if err := run(); err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}
}

func setVersion(debug bool) {
	versionSvc.Name = "Locally API"
	versionSvc.Author = "Carlos Lapao"
	versionSvc.License = "MIT"

	// Use build flag version if available, otherwise fall back to appVersion
	versionToUse := Version
	if versionToUse == "0.0.0" {
		versionToUse = appVersion
	}

	ver, err := version.FromString(versionToUse)
	if err != nil {
		logging.Errorf("Error setting version: %v", err)
	}
	versionSvc.Major = ver.Major
	versionSvc.Minor = ver.Minor
	versionSvc.Build = ver.Build
	versionSvc.Rev = ver.Rev

	// if debug is true, we will load the version from the file as the version is not set
	if debug {
		// loading the version from the file if it exist
		if _, err := os.Stat("../../VERSION"); err == nil {
			versionFile, err := os.ReadFile("../../VERSION")
			if err == nil {
				strVer, err := version.FromString(string(versionFile))
				if err == nil {
					versionSvc.Major = strVer.Major
					versionSvc.Minor = strVer.Minor
					versionSvc.Build = strVer.Build
					versionSvc.Rev = strVer.Rev
				}
			}
		}
	}
}

func initializeSystemService() (system_interfaces.SystemServiceInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing system service...")
	diag := diagnostics.New("initialize_system_service")
	systemService := system.Initialize()
	if systemService == nil {
		diag.AddError("system_service_not_initialized", "system service not initialized", "initialize_system_service", nil)
		return nil, diag
	}
	logging.Info("System service initialized successfully")
	return systemService, diag
}

// initializeDatabase initializes the database service
func initializeDatabase(cfg *config.Config) error {
	logging.Info("Initializing database service...")
	storagePath, err := config.GetInstance().GetStoragePath()
	if err != nil {
		return fmt.Errorf("failed to get storage path: %w", err)
	}
	var dbConfig types.Config
	if cfg.Get(config.DatabaseTypeKey).GetString() == "postgres" {
		dbConfig.Type = types.PostgreSQL
		dbConfig.Host = cfg.Get(config.DatabaseHostKey).GetString()
		dbConfig.Port = cfg.Get(config.DatabasePortKey).GetInt()
		dbConfig.Database = cfg.Get(config.DatabaseDatabaseKey).GetString()
		dbConfig.Username = cfg.Get(config.DatabaseUsernameKey).GetString()
		dbConfig.Password = cfg.Get(config.DatabasePasswordKey).GetString()
		dbConfig.SSLMode = cfg.Get(config.DatabaseSSLModeKey).GetBool()
		if dbConfig.Database == "" {
			return fmt.Errorf("database name is required")
		}
		if dbConfig.Username == "" {
			return fmt.Errorf("database username is required")
		}
		if dbConfig.Password == "" {
			return fmt.Errorf("database password is required")
		}
		if dbConfig.Host == "" {
			return fmt.Errorf("database host is required")
		}
		if dbConfig.Port == 0 {
			dbConfig.Port = 5432
		}
	} else {
		dbConfig.Type = types.SQLite
		dbConfig.StoragePath = storagePath

	}
	dbConfig.Debug = cfg.Get(config.DebugKey).GetBool()

	if err := database.Initialize(&dbConfig); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	logging.Info("Database service initialized successfully")
	return nil
}

// initializeConfigurationStore initializes the configuration store
func initializeConfigurationStore() (stores.ConfigurationDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_configuration_store")
	logging.Info("Initializing configuration store...")
	configurationStore, initDiag := stores.InitializeConfigurationDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Configuration store initialized successfully")
	return configurationStore, diag
}

// initializeApiKeyStore initializes the api key store
func initializeApiKeyStore() (stores.ApiKeyStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing api key store...")
	diag := diagnostics.New("initialize_api_key_store")
	apiKeyStore, initDiag := stores.InitializeApiKeyDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Api key store initialized successfully")
	return apiKeyStore, diag
}

// initializeMessageStore initializes the message store
func initializeMessageStore() (stores.MessageDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing message store...")
	diag := diagnostics.New("initialize_message_store")
	messageStore, initDiag := stores.InitializeMessageDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Message store initialized successfully")
	return messageStore, diag
}

// initializeCertificatesStore initializes the certificates store
func initializeCertificatesStore() (stores.CertificatesDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing certificates store...")
	diag := diagnostics.New("initialize_certificates_store")
	certificatesStore, initDiag := stores.InitializeCertificatesDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Certificates store initialized successfully")
	return certificatesStore, diag
}

// initializeTenantStore initializes the tenant store
func initializeTenantStore() (stores.TenantDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing tenant store...")
	diag := diagnostics.New("initialize_tenant_store")
	tenantStore, initDiag := stores.InitializeTenantDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Tenant store initialized successfully")
	return tenantStore, diag
}

// initializeUserStore initializes the user store
func initializeUserStore() (stores.UserDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing user store...")
	diag := diagnostics.New("initialize_user_store")
	userStore, initDiag := stores.InitializeUserDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("User store initialized successfully")
	return userStore, diag
}

// initializeRoleStore initializes the role store
func initializeRoleStore() (stores.RoleDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing role store...")
	diag := diagnostics.New("initialize_role_store")
	roleStore, initDiag := stores.InitializeRoleDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Role store initialized successfully")
	return roleStore, diag
}

// initializeClaimStore initializes the claim store
func initializeClaimStore() (stores.ClaimDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing claim store...")
	diag := diagnostics.New("initialize_claim_store")
	claimStore, initDiag := stores.InitializeClaimDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Claim store initialized successfully")
	return claimStore, diag
}

// initializeActivityStore initializes the activity store
func initializeActivityStore() (stores.ActivityDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing activity store...")
	diag := diagnostics.New("initialize_activity_store")
	activityStore, initDiag := stores.InitializeActivityDataStore()
	if initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Activity store initialized successfully")
	return activityStore, diag
}

// initializeValidationService initializes the validation service
func initializeValidationService() {
	logging.Info("Initializing validation service...")
	validation.Initialize()
	logging.Info("Validation service initialized successfully")
}

// initializeMessageProcessorService initializes the message processor service
func initializeMessageProcessorService(store stores.MessageDataStoreInterface) (*workers.SystemWorkerMessageService, error) {
	logging.Info("Initializing system messages service...")
	svc, err := workers.Initialize(store)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize system messages service: %w", err)
	}

	logging.Info("System messages service initialized successfully")
	return svc, nil
}

// initializeEncryptionService initializes the encryption service
func initializeEncryptionService(cfg *config.Config) (*encryption.EncryptionService, error) {
	logging.Info("Initializing encryption service...")
	encryptionSvc, err := encryption.Initialize(encryption.Config{
		MasterSecret: cfg.Get(config.EncryptionMasterSecretKey).GetString(),
		GlobalSecret: cfg.Get(config.EncryptionGlobalSecretKey).GetString(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}
	logging.Info("Encryption service initialized successfully")
	return encryptionSvc, nil
}

// initializeAuthService initializes the auth service
func initializeAuthService(cfg *config.Config, authDataStore stores.ApiKeyStoreInterface, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) (auth_interfaces.AuthServiceInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing auth service...")

	authService, diag := auth.Initialize(auth.AuthServiceConfig{
		SecretKey: []byte(cfg.Get(config.JwtAuthSecretKey).GetString()),
		Issuer:    cfg.Get(config.JwtIssuerKey).GetString(),
	}, authDataStore, userStore, tenantStore)

	logging.Info("Auth service initialized successfully")
	return authService, diag
}

// initializeCertificateService initializes the certificate service
func initializeCertificateService(store stores.CertificatesDataStoreInterface) certificates_interfaces.CertificateServiceInterface {
	logging.Info("Initializing certificate service...")
	certificateService := certificates.Initialize(store)
	if certificateService == nil {
		logging.Error("Certificate service not initialized")
		panic("Certificate service not initialized")
	}
	logging.Info("Certificate service initialized successfully")
	return certificateService
}

// initializeTenantService initializes the tenant service
func initializeTenantService(tenantStore stores.TenantDataStoreInterface, userService user_interfaces.UserServiceInterface, roleService rolesvc.RoleServiceInterface, systemService system_interfaces.SystemServiceInterface, claimService claim_interfaces.ClaimServiceInterface) tenant_interfaces.TenantServiceInterface {
	logging.Info("Initializing tenant service...")
	tenantService := tenant.Initialize(tenantStore, userService, roleService, systemService, claimService)
	logging.Info("Tenant service initialized successfully")
	return tenantService
}

// initializeUserService initializes the user service
func initializeUserService(userStore stores.UserDataStoreInterface, roleService rolesvc.RoleServiceInterface, claimService claim_interfaces.ClaimServiceInterface, systemService system_interfaces.SystemServiceInterface) user_interfaces.UserServiceInterface {
	logging.Info("Initializing user service...")
	userService := user.Initialize(userStore, roleService, claimService, systemService)
	logging.Info("User service initialized successfully")
	return userService
}

// initializeRoleService initializes the role service
func initializeRoleService(roleStore stores.RoleDataStoreInterface, systemService system_interfaces.SystemServiceInterface, claimService claim_interfaces.ClaimServiceInterface) rolesvc.RoleServiceInterface {
	logging.Info("Initializing role service...")
	roleService := role.Initialize(roleStore, systemService, claimService)
	logging.Info("Role service initialized successfully")
	return roleService
}

// initializeClaimService initializes the claim service
func initializeClaimService(claimStore stores.ClaimDataStoreInterface) claim_interfaces.ClaimServiceInterface {
	logging.Info("Initializing claim service...")
	claimService := claim.Initialize(claimStore)
	logging.Info("Claim service initialized successfully")
	return claimService
}

// initializeApiKeysService initializes the api keys service
func initializeApiKeysService(apiKeyStore stores.ApiKeyStoreInterface) api_keys_interfaces.ApiKeysServiceInterface {
	logging.Info("Initializing api keys service...")
	apiKeysService := api_keys.Initialize(apiKeyStore)
	logging.Info("Api keys service initialized successfully")
	return apiKeysService
}

// initializeActivityService initializes the activity service
func initializeActivityService(activityStore stores.ActivityDataStoreInterface) activity_interfaces.ActivityServiceInterface {
	logging.Info("Initializing activity service...")
	activityService := activity.Initialize(activityStore)
	logging.Info("Activity service initialized successfully")
	return activityService
}

// initializeAPIServer initializes the API server
func initializeAPIServer(cfg *config.Config, authService auth_interfaces.AuthServiceInterface) (*api.Server, error) {
	logging.Info("Initializing API server...")
	server := api.NewServer(api.Config{
		AuthService: authService,
		Port:        cfg.Get(config.ServerAPIPortKey).GetInt(),
		Hostname:    cfg.Get(config.ServerBindAddressKey).GetString(),
		Prefix:      cfg.Get(config.ServerAPIPrefixKey).GetString(),
	})
	logging.Info("API server initialized successfully")
	return server, nil
}

// initializeEventService initializes the event service for real-time notifications
func initializeEventService() *events.EventService {
	logging.Info("Initializing event service singleton...")
	eventService := events.Initialize() // Initialize the singleton
	logging.Info("Event service singleton initialized successfully")
	return eventService
}

// startEventService starts the event service in the background
func startEventService(ctx *appctx.AppContext, eventService *events.EventService) error {
	logging.Info("Starting event service...")
	if err := eventService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event service: %w", err)
	}
	logging.Info("Event service started successfully")
	return nil
}

func initializeEnvironmentService(ctx *appctx.AppContext, vaults []interfaces.EnvironmentVault) *environment.Environment {
	logging.Info("Initializing environment service...")
	environmentService := environment.Initialize()
	for _, vault := range vaults {
		diag := environmentService.RegisterVault(ctx, vault)
		if diag.HasErrors() {
			logging.Errorf("Error registering vault: %v", diag.GetSummary())
		}
	}
	logging.Info("Environment service initialized successfully")
	return environmentService
}

func seedDatabaseMigrations(ctx *appctx.AppContext, configSvc *config.ConfigService, tenantService tenant_interfaces.TenantServiceInterface, userService user_interfaces.UserServiceInterface, systemService system_interfaces.SystemServiceInterface) *diagnostics.Diagnostics {
	logging.Info("Seeding database migrations...")
	diag := diagnostics.New("seed_database_migrations")

	// Getting dependencies
	db := database.GetInstance()
	if db == nil {
		diag.AddError("database_service_not_initialized", "database service not initialized", "seed_database_migrations", nil)
		return diag
	}
	// Getting the auth store
	authStore := stores.GetApiKeyDataStoreInstance()
	if authStore == nil {
		diag.AddError("auth_store_not_initialized", "auth store not initialized", "seed_database_migrations", nil)
		return diag
	}
	tenantStore := stores.GetTenantDataStoreInstance()
	if tenantStore == nil {
		diag.AddError("tenant_store_not_initialized", "tenant store not initialized", "seed_database_migrations", nil)
		return diag
	}
	userStore := stores.GetUserDataStoreInstance()
	if userStore == nil {
		diag.AddError("user_store_not_initialized", "user store not initialized", "seed_database_migrations", nil)
		return diag
	}

	certificateService := certificates.GetInstance()
	if certificateService == nil {
		diag.AddError("certificate_service_not_initialized", "certificate service not initialized", "seed_database_migrations", nil)
		return diag
	}

	// Initializing the migration service
	service := seeds.NewMigrationService(db.GetDB())
	// migrations workers
	// defaultClaimsMigrationWorker := migrations.NewDefaultClaimsMigrationWorker(db.GetDB(), configSvc.Get())
	// defaultRolesMigrationWorker := migrations.NewDefaultRolesMigrationWorker(db.GetDB(), configSvc.Get())
	defaultTenantMigrationWorker := migrations.NewDefaultTenantMigrationWorker(db.GetDB(), tenantService)
	defaultUsersMigrationWorker := migrations.NewDefaultUsersMigrationWorker(db.GetDB(), configSvc.Get(), systemService, userService)
	rootCertificateMigrationWorker := migrations.NewRootCertificateMigrationWorker(db.GetDB(), certificateService)
	// intermediateCertificateMigrationWorker := migrations.NewIntermediateCertificateMigrationWorker(db.GetDB(), certificateService)

	// service.Register(defaultClaimsMigrationWorker)
	// service.Register(defaultRolesMigrationWorker)
	service.Register(defaultTenantMigrationWorker)
	service.Register(defaultUsersMigrationWorker)
	service.Register(rootCertificateMigrationWorker)
	// service.Register(intermediateCertificateMigrationWorker)

	// Running the migrations
	migrationsDiag := service.RunAll(ctx)
	if migrationsDiag.HasErrors() {
		diag.Append(migrationsDiag)
	} else {
		logging.Info("Database migrations seeded successfully")
	}

	return diag
}

func run() error {
	configSvc, err := config.Initialize()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		return err
	}
	cfg := configSvc.Get()
	setVersion(cfg.GetBool(config.DebugKey, false))
	ctx := appctx.NewContext(context.Background())
	versionSvc.PrintAnsiHeader()

	logging.Initialize()
	logging.Info("Initializing services...")

	// Initializing system service
	systemService, systemServiceDiag := initializeSystemService()
	if systemServiceDiag.HasErrors() {
		return fmt.Errorf("failed to initialize system service: %s", systemServiceDiag.GetSummary())
	}
	// if debug is true, we will print the system service
	if cfg.GetBool(config.DebugKey, false) {
		systemService.LogSummary(ctx)
	}

	// Initializing encryption service
	_, err = initializeEncryptionService(configSvc.Get())
	if err != nil {
		return err
	}

	// Initializing database services
	if err := initializeDatabase(configSvc.Get()); err != nil {
		return err
	}

	// initializing Database Stores
	_, configStoreDiag := initializeConfigurationStore()
	if configStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize configuration store: %s", configStoreDiag.GetSummary())
	}

	apiKeyStore, apiKeyStoreDiag := initializeApiKeyStore()
	if apiKeyStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize api key store: %s", apiKeyStoreDiag.GetSummary())
	}
	messageDataStore, messageStoreDiag := initializeMessageStore()
	if messageStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize message store: %s", messageStoreDiag.GetSummary())
	}
	certificatesStore, certificatesStoreDiag := initializeCertificatesStore()
	if certificatesStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize certificates store: %s", certificatesStoreDiag.GetSummary())
	}

	tenantStore, tenantStoreDiag := initializeTenantStore()
	if tenantStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize tenant store: %s", tenantStoreDiag.GetSummary())
	}

	userStore, userStoreDiag := initializeUserStore()
	if userStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize user store: %s", userStoreDiag.GetSummary())
	}

	// Initialize role store
	roleStore, roleStoreDiag := initializeRoleStore()
	if roleStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize role store: %s", roleStoreDiag.GetSummary())
	}

	// Initialize claim store
	claimStore, claimStoreDiag := initializeClaimStore()
	if claimStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize claim store: %s", claimStoreDiag.GetSummary())
	}

	// Initialize activity store
	activityStore, activityStoreDiag := initializeActivityStore()
	if activityStoreDiag.HasErrors() {
		return fmt.Errorf("failed to initialize activity store: %s", activityStoreDiag.GetSummary())
	}

	// initialize environment service
	vaults := make([]interfaces.EnvironmentVault, 0)

	// Add config vault
	configVault := configvault.New()
	vaults = append(vaults, configVault)

	// Initializing Services

	// initialize environment service
	environmentService := initializeEnvironmentService(ctx, vaults)
	// initializing validation service
	initializeValidationService()
	// initializing event service
	eventService := initializeEventService()
	// Initialize message processor service
	messageService, err := initializeMessageProcessorService(messageDataStore)
	if err != nil {
		return err
	}
	// initialize certificate service
	certificateService := initializeCertificateService(certificatesStore)
	// initialize auth service
	authService, authServiceDiag := initializeAuthService(configSvc.Get(), apiKeyStore, userStore, tenantStore)
	if authServiceDiag.HasErrors() {
		logging.Errorf("Error initializing auth service: %v", authServiceDiag.GetSummary())
		panic(authServiceDiag.GetSummary())
	}

	// initialize claim service
	claimService := initializeClaimService(claimStore)
	// initialize role service
	roleService := initializeRoleService(roleStore, systemService, claimService)
	// initialize user service
	userService := initializeUserService(userStore, roleService, claimService, systemService)
	// initialize tenant service
	tenantService := initializeTenantService(tenantStore, userService, roleService, systemService, claimService)
	// initialize api keys service
	apiKeysService := initializeApiKeysService(apiKeyStore)
	// initialize activity service
	activityService := initializeActivityService(activityStore)
	// initialize API server
	apiServer, err := initializeAPIServer(configSvc.Get(), authService)
	if err != nil {
		return err
	}

	// services initialized, lets start the services default handlers
	logging.Info("Registering routes...")
	// Register Api Default Handlers
	apiServer.RegisterRoutes(api_handlers.NewHandler())
	// Register auth routes
	apiServer.RegisterRoutes(auth_handlers.NewApiHandler(authService, apiKeyStore, activityService))
	// Register event routes using the global singleton
	apiServer.RegisterRoutes(events.NewApiHandler(events.GetInstance(), authService))
	// Register message routes
	apiServer.RegisterRoutes(workers.NewApiHandler(messageService))
	// Register environment routes
	apiServer.RegisterRoutes(environment.NewApiHandler(environmentService))
	// Register certificate routes
	apiServer.RegisterRoutes(certificates.NewApiHandler(certificateService))
	// Register tenant routes
	apiServer.RegisterRoutes(tenant.NewApiHandler(tenantService))
	// Register user routes
	apiServer.RegisterRoutes(user.NewApiHandler(userService))
	// Register role routes
	apiServer.RegisterRoutes(role.NewApiHandler(roleService))
	// Register claim routes
	apiServer.RegisterRoutes(claim.NewApiHandler(claimService))
	// Register api key routes
	apiServer.RegisterRoutes(api_keys.NewApiHandler(apiKeysService))
	// Register activity routes
	apiServer.RegisterRoutes(activity.NewApiHandler(activityService, systemService))
	// Register auth test routes
	apiServer.RegisterRoutes(handlers.NewTestHandler(authService, apiKeyStore, activityService))
	logging.Info("Starting event service...")
	if err := startEventService(ctx, eventService); err != nil {
		return err
	}

	// Registering workers
	logging.Info("Registering message workers...")
	messageService.RegisterWorker(workers.NewEmailWorker())
	messageService.RegisterWorker(workers.NewNotificationWorker())
	messageService.Start(ctx)

	// Seeding database migrations
	seedDiag := seedDatabaseMigrations(ctx, configSvc, tenantService, userService, systemService)
	if seedDiag.HasErrors() {
		panic(seedDiag.GetSummary())
	}

	// Start server in a goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			logging.Errorf("Server error: %v", err)
		}
	}()

	logging.Info("All services started successfully")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown gracefully
	logging.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop event service
	logging.Info("Stopping event service...")
	if err := eventService.Stop(); err != nil {
		logging.Errorf("Error stopping event service: %v", err)
	} else {
		logging.Info("Event service stopped successfully")
	}

	// Stop API server
	logging.Info("Stopping API server...")
	if err := apiServer.Stop(shutdownCtx); err != nil {
		logging.Errorf("Error shutting down server: %v", err)
		return fmt.Errorf("error shutting down server: %w", err)
	}

	logging.Info("Application shutdown completed successfully")
	return nil
}
