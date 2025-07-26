package entities

import "github.com/cjlapao/locally-cli/pkg/models"

type Claim struct {
	BaseModel
	Service string             `json:"service" gorm:"not null;type:text"`
	Module  string             `json:"module" gorm:"not null;type:text"`
	Action  models.ClaimAction `json:"action" gorm:"not null;type:text"`
}

func (Claim) TableName() string {
	return "claims"
}
