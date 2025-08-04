package service

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

func MapCreateUserRequestToEntity(user *models.CreateUserRequest) *entities.User {
	result := &entities.User{
		BaseModel: entities.BaseModel{
			ID: user.ID,
		},
		Name:     user.Name,
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
	}

	result.Slug = utils.Slugify(user.Name)
	result.Status = "active"
	result.Blocked = false
	result.TwoFactorEnabled = user.TwoFactorEnabled
	result.TwoFactorSecret = ""
	result.TwoFactorVerified = false
	result.RefreshToken = ""
	result.Roles = []entities.Role{}
	result.Claims = []entities.Claim{}

	return result
}

func MapUpdateUserRequestToEntity(user *models.UpdateUserRequest) *entities.User {
	result := &entities.User{
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	}

	return result
}
