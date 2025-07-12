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
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/encryption"
	"github.com/cjlapao/locally-cli/internal/events"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/internal/messages"
	"github.com/cjlapao/locally-cli/internal/validation"
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
	var dbConfig database.Config
	if cfg.Get(config.DatabaseTypeKey).GetString() == "postgres" {
		dbConfig.Type = database.PostgreSQL
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
		dbConfig.Type = database.SQLite
		dbConfig.StoragePath = storagePath

	}
	dbConfig.Debug = cfg.Get(config.DebugKey).GetBool()

	if err := database.Initialize(&dbConfig); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	logging.Info("Database service initialized successfully")
	return nil
}

// initializeAuthStore initializes the auth store
func initializeAuthStore() (*stores.AuthDataStore, error) {
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

// initializeValidationService initializes the validation service
func initializeValidationService() {
	logging.Info("Initializing validation service...")
	validation.Initialize()
	logging.Info("Validation service initialized successfully")
}

// initializeMessageProcessorService initializes the message processor service
func initializeMessageProcessorService(store *stores.MessageDataStore) (*messages.SystemMessageService, error) {
	logging.Info("Initializing system messages service...")
	svc, err := messages.Initialize(store)
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
func initializeAuthService(cfg *config.Config, authDataStore *stores.AuthDataStore) *auth.AuthService {
	logging.Info("Initializing auth service...")

	authService := auth.NewService(auth.AuthServiceConfig{
		SecretKey: cfg.Get(config.JwtAuthSecretKey).GetString(),
	}, authDataStore)
	logging.Info("Auth service initialized successfully")
	return authService
}

// initializeAPIServer initializes the API server
func initializeAPIServer(cfg *config.Config, authService *auth.AuthService) (*api.Server, error) {
	logging.Info("Initializing API server...")
	server := api.NewServer(api.Config{
		Port:           cfg.Get(config.ServerAPIPortKey).GetInt(),
		Hostname:       cfg.Get(config.ServerBindAddressKey).GetString(),
		Prefix:         cfg.Get(config.ServerAPIPrefixKey).GetString(),
		AuthMiddleware: auth.NewRequireAuthMiddleware(authService),
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
func startEventService(eventService *events.EventService, ctx api.ApiContext) error {
	logging.Info("Starting event service...")
	if err := eventService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event service: %w", err)
	}
	logging.Info("Event service started successfully")
	return nil
}

func run() error {
	setVersion()
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
	authDataStore, err := initializeAuthStore()
	if err != nil {
		return err
	}
	messageDataStore, err := initializeMessageStore()
	if err != nil {
		return err
	}

	// initializing validation service
	initializeValidationService()

	// initializing event service
	eventService := initializeEventService()

	// Initialize message processor service
	messageService, err := initializeMessageProcessorService(messageDataStore)
	if err != nil {
		return err
	}

	// initialize auth service
	authService := initializeAuthService(configSvc.Get(), authDataStore)

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
	apiServer.RegisterRoutes(auth.NewApiHandler(authService, authDataStore))
	// Register event routes using the global singleton
	apiServer.RegisterRoutes(events.NewApiHandler(events.GetInstance(), authService))
	// Register message routes
	apiServer.RegisterRoutes(messages.NewApiHandler(messageService))

	logging.Info("Starting event service...")
	ctx := api.NewContext(context.Background())
	// Start event service
	if err := startEventService(eventService, *ctx); err != nil {
		return err
	}

	// Registering workers
	logging.Info("Registering message workers...")
	messageService.RegisterWorker(messages.NewEmailWorker())
	messageService.RegisterWorker(messages.NewNotificationWorker())
	messageService.Start(*ctx)

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
