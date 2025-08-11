package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	user_models "github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

func MapUserToDto(user *entities.User) *models.User {
	result := &models.User{
		ID:                    user.ID,
		Slug:                  user.Slug,
		Name:                  user.Name,
		Username:              user.Username,
		Password:              user.Password,
		Email:                 user.Email,
		Status:                user.Status,
		TenantID:              user.TenantID,
		TwoFactorEnabled:      user.TwoFactorEnabled,
		TwoFactorSecret:       user.TwoFactorSecret,
		TwoFactorVerified:     user.TwoFactorVerified,
		Blocked:               user.Blocked,
		RefreshToken:          user.RefreshToken,
		RefreshTokenExpiresAt: user.RefreshTokenExpiresAt,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}

	// Obfuscate the password to avoid leaking it in the response
	result.Password = utils.ObfuscateString(result.Password)

	// Map the roles
	roles := []models.Role{}
	for _, role := range user.Roles {
		role := MapRoleToDto(&role)
		roles = append(roles, *role)
	}
	result.Roles = roles

	// Map the claims
	claims := []models.Claim{}
	for _, claim := range user.Claims {
		claim := MapClaimToDto(&claim)
		claims = append(claims, *claim)
	}
	result.Claims = claims

	return result
}

func MapUsersToDto(users []entities.User) []models.User {
	result := make([]models.User, len(users))
	for i, user := range users {
		result[i] = *MapUserToDto(&user)
	}
	return result
}

func MapSelfEntityUserToDto(user *entities.User) *user_models.SelfUserResponse {
	result := &user_models.SelfUserResponse{
		ID:               user.ID,
		Slug:             user.Slug,
		Name:             user.Name,
		Username:         user.Username,
		Email:            user.Email,
		Status:           user.Status,
		TwoFactorEnabled: user.TwoFactorEnabled,
	}

	// Map the roles
	roles := []string{}
	for _, role := range user.Roles {
		roles = append(roles, role.Slug)
	}
	result.Roles = roles

	// Map the claims
	claims := []string{}
	for _, claim := range user.Claims {
		claims = append(claims, claim.Slug)
	}
	result.Claims = claims

	return result
}

func MapSelfUserToDto(user *models.User) *user_models.SelfUserResponse {
	result := &user_models.SelfUserResponse{
		ID:               user.ID,
		Slug:             user.Slug,
		Name:             user.Name,
		Username:         user.Username,
		Email:            user.Email,
		Status:           user.Status,
		TwoFactorEnabled: user.TwoFactorEnabled,
	}

	// Map the roles
	roles := []string{}
	for _, role := range user.Roles {
		roles = append(roles, role.Slug)
	}
	result.Roles = roles

	// Map the claims
	claims := []string{}
	for _, claim := range user.Claims {
		claims = append(claims, claim.Slug)
	}
	result.Claims = claims

	return result
}
