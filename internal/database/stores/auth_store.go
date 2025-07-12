package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/encryption"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	authDataStoreInstance *AuthDataStore
	authDataStoreOnce     sync.Once
)

// AuthDataStore handles auth-specific database operations
type AuthDataStore struct {
	database.BaseDataStore
}

// GetAuthDataStoreInstance returns the singleton instance of the auth store
func GetAuthDataStoreInstance() *AuthDataStore {
	return authDataStoreInstance
}

// InitializeAuthDataStore initializes the auth store singleton
func InitializeAuthDataStore() error {
	var initErr error
	cfg := config.GetInstance().Get()
	authDataStoreOnce.Do(func() {
		// Get the database service instance
		dbService := database.GetInstance()
		if dbService == nil {
			initErr = fmt.Errorf("database service not initialized")
			return
		}

		store := &AuthDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running auth migrations")
			if err := store.Migrate(); err != nil {
				initErr = fmt.Errorf("failed to run auth migrations: %w", err)
				return
			}
			logging.Info("Auth migrations completed")
		}

		authDataStoreInstance = store
	})

	return initErr
}

// Migrate implements the DataStore interface
func (s *AuthDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&types.User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}

	err := s.createDefaultRootUser()
	if err != nil {
		return fmt.Errorf("failed to create default root user: %w", err)
	}

	return nil
}

func (s *AuthDataStore) createDefaultRootUser() error {
	cfg := config.GetInstance().Get()
	currentEnvironmentRootPassword := cfg.Get(config.RootUserPasswordKey).GetString()
	currentDBRootUser, err := s.GetUserByUsername(context.Background(), "root")
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check if root user already exists: %w", err)
		}
	}
	if currentDBRootUser != nil {
		logging.Info("Root user already exists, checking if password needs to be updated")
		if currentEnvironmentRootPassword != "" {
			// let's check if the password is the default password
			encryptionService := encryption.GetInstance()
			if err := encryptionService.VerifyPassword(currentEnvironmentRootPassword, currentDBRootUser.Password); err != nil {
				logging.Info("Password is not the default password, updating it")
				// if the password is not the default password, we need to update it
				currentDBRootUser.Password = currentEnvironmentRootPassword
				err = s.UpdateUser(context.Background(), currentDBRootUser)
				if err != nil {
					return fmt.Errorf("failed to update root user password: %w", err)
				}
			}
			logging.Info("Root user password updated")
			return nil
		}

		return nil
	}

	if currentEnvironmentRootPassword == "" {
		return fmt.Errorf("root user password is empty")
	}

	user := &types.User{
		Name:     "Root",
		Username: "root",
		Email:    "root@parallels.com",
		Password: currentEnvironmentRootPassword,
		Role:     "su",
		Status:   "active",
		Blocked:  false,
	}

	dbUser, err := s.CreateUser(context.Background(), user)
	if err != nil {
		return fmt.Errorf("failed to create default root user: %w", err)
	}

	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt

	logging.Info("Default root user created successfully")
	return nil
}

// CreateUser creates a new user
func (s *AuthDataStore) CreateUser(ctx context.Context, user *types.User) (*types.User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	user.Password = encryptedPassword

	result := s.GetDB().Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthDataStore) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	result := s.GetDB().First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", result.Error)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *AuthDataStore) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	var user types.User
	result := s.GetDB().First(&user, "username = ?", username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (s *AuthDataStore) UpdateUser(ctx context.Context, user *types.User) error {
	user.UpdatedAt = time.Now()
	currentUser, err := s.GetUserByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	if user.Password != currentUser.Password {
		encryptionService := encryption.GetInstance()
		encryptedPassword, err := encryptionService.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		user.Password = encryptedPassword
	}
	if user.Name != currentUser.Name {
		user.Name = currentUser.Name
	}
	if user.Email != currentUser.Email {
		user.Email = currentUser.Email
	}
	if user.Role != currentUser.Role {
		user.Role = currentUser.Role
	}
	if user.Status != currentUser.Status {
		user.Status = currentUser.Status
	}
	if user.Blocked != currentUser.Blocked {
		user.Blocked = currentUser.Blocked
	}
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(currentUser).Error
}

func (s *AuthDataStore) UpdateUserPassword(ctx context.Context, id string, password string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	user.Password = encryptedPassword
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

func (s *AuthDataStore) BlockUser(ctx context.Context, id string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.Blocked = true
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

func (s *AuthDataStore) SetRefreshToken(ctx context.Context, id string, refreshToken string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.RefreshToken = refreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(24 * time.Hour)
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

// DeleteUser deletes a user
func (s *AuthDataStore) DeleteUser(ctx context.Context, id string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	return s.GetDB().Delete(user).Error
}
