package mappers

import (
	db_entities "github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapBaseModelToDto(baseModel *db_entities.BaseModel) *models.BaseModel {
	return &models.BaseModel{
		ID:        baseModel.ID,
		Slug:      baseModel.Slug,
		CreatedBy: baseModel.CreatedBy,
		UpdatedBy: baseModel.UpdatedBy,
		CreatedAt: baseModel.CreatedAt,
		UpdatedAt: baseModel.UpdatedAt,
	}
}

func MapBaseModelWithTenantToDto(baseModel *db_entities.BaseModelWithTenant) *models.BaseModelWithTenant {
	return &models.BaseModelWithTenant{
		ID:        baseModel.ID,
		TenantID:  baseModel.TenantID,
		Slug:      baseModel.Slug,
		CreatedBy: baseModel.CreatedBy,
		UpdatedBy: baseModel.UpdatedBy,
		CreatedAt: baseModel.CreatedAt,
		UpdatedAt: baseModel.UpdatedAt,
	}
}
