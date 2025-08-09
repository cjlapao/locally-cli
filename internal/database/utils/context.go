package utils

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
)

func GetBaseModelFromContext(ctx *appctx.AppContext) *entities.BaseModel {
	result := &entities.BaseModel{}
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	result.CreatedBy = ctx.GetUserID()
	result.UpdatedBy = ctx.GetUserID()
	return result
}

func GetTenantBaseModelFromContext(ctx *appctx.AppContext) *entities.BaseModelWithTenant {
	result := &entities.BaseModelWithTenant{}
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	result.TenantID = ctx.GetTenantID()
	result.CreatedBy = ctx.GetUserID()
	result.UpdatedBy = ctx.GetUserID()
	return result
}
