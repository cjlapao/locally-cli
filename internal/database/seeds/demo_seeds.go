// Package seeds implements a service that seeds the database with demo data.
package seeds

import (
	"context"
	"fmt"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/logging"
)

func createSuperUser(ctx context.Context, authDataStore *stores.AuthDataStore) (*types.User, error) {
	// check if the user already exists
	var user *types.User
	var err error

	user, err = authDataStore.GetUserByUsername(ctx, "parallels")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		logging.WithField("username", "parallels").Info("User not found, creating user")
		// create a new user
		user = &types.User{
			TenantID: config.GlobalTenantID,
			Username: "parallels",
			Password: "parallels",
			Email:    "parallels@parallelsdev.com",
			Name:     "Parallels Development User",
			Role:     "admin",
			Status:   "active",
		}

		dbUser, err := authDataStore.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		logging.WithField("username", user.Username).Info("User created")
		user.ID = dbUser.ID
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
	}

	return user, nil
}
