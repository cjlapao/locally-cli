package mappers

import (
	db_entities "github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapBaseModelToEntity(baseModel *models.BaseModel) *db_entities.BaseModel {
	return &db_entities.BaseModel{
		ID:        baseModel.ID,
		Slug:      baseModel.Slug,
		CreatedBy: baseModel.CreatedBy,
		UpdatedBy: baseModel.UpdatedBy,
		CreatedAt: baseModel.CreatedAt,
		UpdatedAt: baseModel.UpdatedAt,
	}
}

func MapBaseModelWithTenantToEntity(baseModel *models.BaseModelWithTenant) *db_entities.BaseModelWithTenant {
	return &db_entities.BaseModelWithTenant{
		ID:        baseModel.ID,
		TenantID:  baseModel.TenantID,
		Slug:      baseModel.Slug,
		CreatedBy: baseModel.CreatedBy,
		UpdatedBy: baseModel.UpdatedBy,
		CreatedAt: baseModel.CreatedAt,
		UpdatedAt: baseModel.UpdatedAt,
	}
}
