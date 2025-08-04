package entities

import "github.com/cjlapao/locally-cli/pkg/models"

type Role struct {
	BaseModelWithTenant
	Name          string               `json:"name" gorm:"not null;type:text"`
	Description   string               `json:"description" gorm:"not null;type:text"`
	SecurityLevel models.SecurityLevel `json:"security_level" gorm:"not null;type:text"`
	Claims        []Claim              `json:"claims" gorm:"many2many:role_claims;"`
}

func (Role) TableName() string {
	return "roles"
}
