package utils

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/google/uuid"
)

func GetBaseModelFromContext(ctx *appctx.AppContext, baseModel *entities.BaseModel) *entities.BaseModel {
	if baseModel == nil {
		baseModel = &entities.BaseModel{}
	}
	baseModel.ID = uuid.New().String()
	baseModel.CreatedAt = time.Now()
	baseModel.UpdatedAt = time.Now()
	baseModel.CreatedBy = ctx.GetUserID()
	baseModel.UpdatedBy = ctx.GetUserID()
	return baseModel
}

func GetTenantBaseModelFromContext(ctx *appctx.AppContext, baseModel *entities.BaseModelWithTenant) *entities.BaseModelWithTenant {
	if baseModel == nil {
		baseModel = &entities.BaseModelWithTenant{}
	}
	baseModel.ID = uuid.New().String()
	baseModel.CreatedAt = time.Now()
	baseModel.UpdatedAt = time.Now()
	baseModel.TenantID = ctx.GetTenantID()
	baseModel.CreatedBy = ctx.GetUserID()
	baseModel.UpdatedBy = ctx.GetUserID()
	return baseModel
}
