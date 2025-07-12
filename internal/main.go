package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cjlapao/lxc-agent/internal/api"
	"github.com/cjlapao/lxc-agent/internal/auth"
	"github.com/cjlapao/lxc-agent/internal/cache"
	"github.com/cjlapao/lxc-agent/internal/capsule"
	"github.com/cjlapao/lxc-agent/internal/config"
	"github.com/cjlapao/lxc-agent/internal/database"
	"github.com/cjlapao/lxc-agent/internal/database/stores"
	"github.com/cjlapao/lxc-agent/internal/encryption"
	"github.com/cjlapao/lxc-agent/internal/events"
	"github.com/cjlapao/lxc-agent/internal/logging"
	"github.com/cjlapao/lxc-agent/internal/message_processor"
	"github.com/cjlapao/lxc-agent/internal/validation"
)

const (
	// Version is the version of the application
	Version = "1.0.0"
	// AppName is the name of the application
	AppName = "Capsule Registry"
)

func main() {
	// Initialize configuration service first
	if err := config.Initialize(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging service
	logging.Initialize()
	logging.Info("Starting Capsule Registry...")

	// Define command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
	)

	// Parse command line arguments
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", AppName, Version)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		showUsage()
		os.Exit(0)
	}

	cfg := config.GetInstance().Get()

	// Initialize services
	if err := run(cfg); err != nil {
		logging.Errorf("Error: %v", err)
		os.Exit(1)
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

// initializeCapsuleBlueprintStore initializes the capsule blueprint store
func initializeCapsuleBlueprintStore() (*stores.CapsuleBlueprintDataStore, error) {
	logging.Info("Initializing capsule blueprint store...")
	if err := stores.InitializeBlueprintCapsuleDataStore(); err != nil {
		return nil, fmt.Errorf("failed to initialize capsule blueprint store: %w", err)
	}
	logging.Info("Capsule blueprint store initialized successfully")
	return stores.GetCapsuleBlueprintDataStoreInstance(), nil
}

// initializeValidationService initializes the validation service
func initializeValidationService() {
	logging.Info("Initializing validation service...")
	validation.Initialize()
	logging.Info("Validation service initialized successfully")
}

// initializeCacheService initializes the cache service
func initializeCacheService() error {
	logging.Info("Initializing cache service...")
	if err := cache.Initialize(cache.Config{
		CleanupInterval: 5 * time.Minute,
	}); err != nil {
		return fmt.Errorf("failed to initialize cache service: %w", err)
	}
	logging.Info("Cache service initialized successfully")
	return nil
}

// initializeMessageProcessorService initializes the message processor service
func initializeMessageProcessorService(store *stores.MessageDataStore) (*message_processor.MessageProcessorService, error) {
	logging.Info("Initializing message processor service...")
	svc, err := message_processor.Initialize(store)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize message processor service: %w", err)
	}

	logging.Info("Message processor service initialized successfully")
	return svc, nil
}

// initializeEncryptionService initializes the encryption service
func initializeEncryptionService(cfg *config.Config) error {
	logging.Info("Initializing encryption service...")
	if err := encryption.Initialize(encryption.Config{
		MasterSecret: cfg.Get(config.EncryptionMasterSecretKey).GetString(),
		GlobalSecret: cfg.Get(config.EncryptionGlobalSecretKey).GetString(),
	}); err != nil {
		return fmt.Errorf("failed to initialize encryption service: %w", err)
	}
	logging.Info("Encryption service initialized successfully")
	return nil
}

// initializeAuthService initializes the auth service
func initializeAuthService(cfg *config.Config, authDataStore *stores.AuthDataStore) auth.Service {
	logging.Info("Initializing auth service...")

	authService := auth.NewService(auth.Config{
		SecretKey: cfg.Get(config.JwtAuthSecretKey).GetString(),
	}, authDataStore)
	logging.Info("Auth service initialized successfully")
	return authService
}

// initializeAPIServer initializes the API server
func initializeAPIServer(cfg *config.Config, authService auth.Service) (*api.Server, error) {
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
func initializeEventService() error {
	logging.Info("Initializing event service singleton...")
	events.Initialize() // Initialize the singleton
	logging.Info("Event service singleton initialized successfully")
	return nil
}

// startEventService starts the event service in the background
func startEventService(ctx context.Context) error {
	logging.Info("Starting event service...")
	eventService := events.GetGlobalService()
	if err := eventService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event service: %w", err)
	}
	logging.Info("Event service started successfully")
	return nil
}

func run(cfg *config.Config) error {
	logging.Info("Initializing application...")

	if err := initializeEncryptionService(cfg); err != nil {
		return err
	}

	// Initializing database services
	if err := initializeDatabase(cfg); err != nil {
		return err
	}

	authDataStore, err := initializeAuthStore()
	if err != nil {
		return err
	}

	messageDataStore, err := initializeMessageStore()
	if err != nil {
		return err
	}

	initializeValidationService()

	if err := initializeCacheService(); err != nil {
		return err
	}

	// Initialize event service singleton
	if err := initializeEventService(); err != nil {
		return err
	}

	// Initialize message processor service
	messageProcessorService, err := initializeMessageProcessorService(messageDataStore)
	if err != nil {
		return err
	}

	// Initialize auth service
	authService := initializeAuthService(cfg, authDataStore)

	// Initialize capsule store
	capsuleDataStore, err := initializeCapsuleBlueprintStore()
	if err != nil {
		return err
	}

	// Initialize API server
	server, err := initializeAPIServer(cfg, authService)
	if err != nil {
		return err
	}

	logging.Info("Registering routes...")
	// Register health check routes
	server.RegisterRoutes(api.NewHandler())
	// Register auth routes
	server.RegisterRoutes(auth.NewApiHandler(authService, authDataStore))
	// Register event routes using the global singleton
	server.RegisterRoutes(events.NewApiHandler(events.GetGlobalService(), authService))
	// Register message routes
	server.RegisterRoutes(message_processor.NewApiHandler(message_processor.GetInstance()))
	// Register capsule blueprint routes
	server.RegisterRoutes(capsule.NewBlueprintApiHandler(capsuleDataStore))

	ctx := context.Background()
	// Start event service
	if err := startEventService(ctx); err != nil {
		return err
	}

	// TODO: Create initial test messages if in debug mode
	// if cfg.Get(config.DebugKey).GetBool() {
	//
	// }

	// TODO: Seed demo data
	// if cfg.Get(config.SeedDemoDataKey).GetBool() {
	//
	// }

	// Registering workers
	messageProcessorService.RegisterWorker(message_processor.NewEmailWorker())
	messageProcessorService.RegisterWorker(message_processor.NewNotificationWorker())
	messageProcessorService.Start(ctx)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
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
	if err := events.GetGlobalService().Stop(); err != nil {
		logging.Errorf("Error stopping event service: %v", err)
	} else {
		logging.Info("Event service stopped successfully")
	}

	// Stop API server
	logging.Info("Stopping API server...")
	if err := server.Stop(shutdownCtx); err != nil {
		logging.Errorf("Error shutting down server: %v", err)
		return fmt.Errorf("error shutting down server: %w", err)
	}

	logging.Info("Application shutdown completed successfully")
	return nil
}

func showUsage() {
	fmt.Printf("%s - A command line tool for container management\n\n", AppName)
	fmt.Println("Usage:")
	fmt.Printf("  %s [options]\n\n", AppName)
	fmt.Println("Options:")
	fmt.Println("  --help              Show this help message")
	fmt.Println("  --version           Show version information")
	fmt.Println("  --config <path>     Path to configuration file (JSON or YAML)")
	fmt.Println("  --port <port>       Port to run the API server on")
	fmt.Println("  --hostname <host>   Hostname to run the API server on")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println()
	fmt.Println("Configuration file formats supported: JSON, YAML")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s --version\n", AppName)
	fmt.Printf("  %s --config config.yaml\n", AppName)
	fmt.Printf("  %s --username admin --password secret\n", AppName)
}
