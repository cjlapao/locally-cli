package entities

import (
	"fmt"

	"github.com/cjlapao/locally-cli/pkg/models"
)

type Claim struct {
	BaseModelWithTenant
	Service       string               `json:"service" gorm:"not null;type:text"`
	Module        string               `json:"module" gorm:"not null;type:text"`
	Action        models.AccessLevel   `json:"action" gorm:"not null;type:text"`
	SecurityLevel models.SecurityLevel `json:"security_level" gorm:"not null;type:text"`
}

func (c *Claim) GetSlug() string {
	return fmt.Sprintf("%s::%s::%s", c.Service, c.Module, c.Action)
}

func (Claim) TableName() string {
	return "claims"
}
