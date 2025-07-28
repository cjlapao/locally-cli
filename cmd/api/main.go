package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cjlapao/common-go/version"
	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/auth/handlers"
	"github.com/cjlapao/locally-cli/internal/certificates"
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
	"github.com/cjlapao/locally-cli/internal/tenant"
	"github.com/cjlapao/locally-cli/internal/user"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/internal/vaults/configvault"
	"github.com/cjlapao/locally-cli/internal/workers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/interfaces"
)

var versionSvc = version.Get()

func main() {
	logging.Info("Starting locally API")
	if err := run(); err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}
}

func setVersion() {
	versionSvc.Name = "Locally API"
	versionSvc.Author = "Carlos Lapao"
	versionSvc.License = "MIT"

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
func initializeConfigurationStore() (*stores.ConfigurationDataStore, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_configuration_store")
	logging.Info("Initializing configuration store...")
	if initDiag := stores.InitializeConfigurationDataStore(); initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Configuration store initialized successfully")
	return stores.GetConfigurationDataStoreInstance(), diag
}

// initializeAuthStore initializes the auth store
func initializeAuthStore() (stores.AuthDataStoreInterface, error) {
	logging.Info("Initializing auth store...")
	if err := stores.InitializeAuthDataStore(); err != nil {
		return nil, fmt.Errorf("failed to initialize auth store: %w", err)
	}
	logging.Info("Auth store initialized successfully")
	return stores.GetAuthDataStoreInstance(), nil
}

// initializeMessageStore initializes the message store
func initializeMessageStore() (*stores.MessageDataStore, error) {
	logging.Info("Initializing message store...")
	if err := stores.InitializeMessageDataStore(); err != nil {
		return nil, fmt.Errorf("failed to initialize message store: %w", err)
	}
	logging.Info("Message store initialized successfully")
	return stores.GetMessageDataStoreInstance(), nil
}

// initializeCertificatesStore initializes the certificates store
func initializeCertificatesStore() (*stores.CertificatesDataStore, *diagnostics.Diagnostics) {
	logging.Info("Initializing certificates store...")
	diag := diagnostics.New("initialize_certificates_store")
	if initDiag := stores.InitializeCertificatesDataStore(); initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Certificates store initialized successfully")
	return stores.GetCertificatesDataStoreInstance(), diag
}

// initializeTenantStore initializes the tenant store
func initializeTenantStore() (stores.TenantDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing tenant store...")
	diag := diagnostics.New("initialize_tenant_store")
	if initDiag := stores.InitializeTenantDataStore(); initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("Tenant store initialized successfully")
	return stores.GetTenantDataStoreInstance(), diag
}

// initializeUserStore initializes the user store
func initializeUserStore() (stores.UserDataStoreInterface, *diagnostics.Diagnostics) {
	logging.Info("Initializing user store...")
	diag := diagnostics.New("initialize_user_store")
	if initDiag := stores.InitializeUserDataStore(); initDiag.HasErrors() {
		diag.Append(initDiag)
		return nil, diag
	}
	logging.Info("User store initialized successfully")
	return stores.GetUserDataStoreInstance(), diag
}

// initializeValidationService initializes the validation service
func initializeValidationService() {
	logging.Info("Initializing validation service...")
	validation.Initialize()
	logging.Info("Validation service initialized successfully")
}

// initializeMessageProcessorService initializes the message processor service
func initializeMessageProcessorService(store *stores.MessageDataStore) (*workers.SystemWorkerMessageService, error) {
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
func initializeAuthService(cfg *config.Config, authDataStore stores.AuthDataStoreInterface, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) (*auth.AuthService, *diagnostics.Diagnostics) {
	logging.Info("Initializing auth service...")

	authService, diag := auth.Initialize(auth.AuthServiceConfig{
		SecretKey: []byte(cfg.Get(config.JwtAuthSecretKey).GetString()),
		Issuer:    cfg.Get(config.JwtIssuerKey).GetString(),
	}, authDataStore, userStore, tenantStore)

	logging.Info("Auth service initialized successfully")
	return authService, diag
}

// initializeCertificateService initializes the certificate service
func initializeCertificateService(store *stores.CertificatesDataStore) *certificates.CertificateService {
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
func initializeTenantService(tenantStore stores.TenantDataStoreInterface) tenant.TenantServiceInterface {
	logging.Info("Initializing tenant service...")
	tenantService := tenant.Initialize(tenantStore)
	logging.Info("Tenant service initialized successfully")
	return tenantService
}

// initializeUserService initializes the user service
func initializeUserService(userStore stores.UserDataStoreInterface) user.UserServiceInterface {
	logging.Info("Initializing user service...")
	userService := user.Initialize(userStore)
	logging.Info("User service initialized successfully")
	return userService
}

// initializeAPIServer initializes the API server
func initializeAPIServer(cfg *config.Config, authService *auth.AuthService) (*api.Server, error) {
	logging.Info("Initializing API server...")
	server := api.NewServer(api.Config{
		Port:                cfg.Get(config.ServerAPIPortKey).GetInt(),
		Hostname:            cfg.Get(config.ServerBindAddressKey).GetString(),
		Prefix:              cfg.Get(config.ServerAPIPrefixKey).GetString(),
		AuthMiddleware:      api.NewRequireAuthPreMiddleware(authService),
		SuperUserMiddleware: api.NewRequireSuperUserPreMiddleware(authService),
	}, nil)
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

func seedDatabaseMigrations(ctx *appctx.AppContext, configSvc *config.ConfigService) *diagnostics.Diagnostics {
	logging.Info("Seeding database migrations...")
	diag := diagnostics.New("seed_database_migrations")

	// Getting dependencies
	db := database.GetInstance()
	if db == nil {
		diag.AddError("database_service_not_initialized", "database service not initialized", "seed_database_migrations", nil)
		return diag
	}
	// Getting the auth store
	authStore := stores.GetAuthDataStoreInstance()
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
	defaultClaimsMigrationWorker := migrations.NewDefaultClaimsMigrationWorker(db.GetDB(), configSvc.Get())
	defaultRolesMigrationWorker := migrations.NewDefaultRolesMigrationWorker(db.GetDB(), configSvc.Get())
	defaultTenantMigrationWorker := migrations.NewDefaultTenantMigrationWorker(db.GetDB(), tenantStore)
	defaultUsersMigrationWorker := migrations.NewDefaultUsersMigrationWorker(db.GetDB(), configSvc.Get(), userStore, tenantStore)
	rootCertificateMigrationWorker := migrations.NewRootCertificateMigrationWorker(db.GetDB(), certificateService)
	intermediateCertificateMigrationWorker := migrations.NewIntermediateCertificateMigrationWorker(db.GetDB(), certificateService)

	service.Register(defaultClaimsMigrationWorker)
	service.Register(defaultRolesMigrationWorker)
	service.Register(defaultTenantMigrationWorker)
	service.Register(defaultUsersMigrationWorker)
	service.Register(rootCertificateMigrationWorker)
	service.Register(intermediateCertificateMigrationWorker)

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
	setVersion()
	ctx := appctx.NewContext(context.Background())
	versionSvc.PrintAnsiHeader()
	configSvc, err := config.Initialize()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		return err
	}

	logging.Initialize()
	logging.Info("Initializing services...")

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

	authDataStore, err := initializeAuthStore()
	if err != nil {
		return err
	}
	messageDataStore, err := initializeMessageStore()
	if err != nil {
		return err
	}

	certificatesStore, certificatesStoreDiag := initializeCertificatesStore()
	if certificatesStoreDiag.HasErrors() {
		return err
	}

	tenantStore, tenantStoreDiag := initializeTenantStore()
	if tenantStoreDiag.HasErrors() {
		return err
	}

	userStore, userStoreDiag := initializeUserStore()
	if userStoreDiag.HasErrors() {
		return err
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
	authService, authServiceDiag := initializeAuthService(configSvc.Get(), authDataStore, userStore, tenantStore)
	if authServiceDiag.HasErrors() {
		logging.Errorf("Error initializing auth service: %v", authServiceDiag.GetSummary())
		panic(authServiceDiag.GetSummary())
	}
	// initialize tenant service
	tenantService := initializeTenantService(tenantStore)
	// initialize user service
	userService := initializeUserService(userStore)
	// initialize API server
	apiServer, err := initializeAPIServer(configSvc.Get(), authService)
	if err != nil {
		return err
	}

	// services initialized, lets start the services default handlers
	logging.Info("Registering routes...")
	// Register health check routes
	apiServer.RegisterRoutes(api.NewHandler())
	// Register auth routes
	apiServer.RegisterRoutes(handlers.NewApiHandler(authService, authDataStore))
	// Register event routes using the global singleton
	apiServer.RegisterRoutes(events.NewApiHandler(events.GetInstance(), authService))
	// Register message routes
	apiServer.RegisterRoutes(workers.NewApiHandler(messageService))
	// Register environment routes
	apiServer.RegisterRoutes(environment.NewApiHandler(environmentService))
	// Register certificate routes
	apiServer.RegisterRoutes(certificates.NewApiHandlers(certificateService, certificatesStore))
	// Register tenant routes
	apiServer.RegisterRoutes(tenant.NewApiHandler(tenantService))
	// Register user routes
	apiServer.RegisterRoutes(user.NewApiHandler(userService))
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
	seedDiag := seedDatabaseMigrations(ctx, configSvc)
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
