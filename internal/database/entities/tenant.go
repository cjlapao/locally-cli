package entities

import "time"

type Tenant struct {
	BaseModel
	Name          string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Description   string     `json:"description" gorm:"column:description;type:text"`
	Domain        string     `json:"domain" gorm:"column:domain;type:varchar(255);not null;unique"`
	OwnerID       string     `json:"owner_id" gorm:"column:owner_id;type:varchar(255);"`
	ContactEmail  string     `json:"contact_email" gorm:"column:contact_email;type:varchar(255);"`
	Status        string     `json:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	ActivatedAt   *time.Time `json:"activated_at" gorm:"column:activated_at;type:timestamp;"`
	DeactivatedAt *time.Time `json:"deactivated_at" gorm:"column:deactivated_at;type:timestamp;"`
	Metadata      string     `json:"metadata" gorm:"column:metadata;type:json;"`
	LogoURL       string     `json:"logo_url" gorm:"column:logo_url;type:varchar(255);"`
	Require2FA    bool       `json:"require_2fa" gorm:"column:require_2fa;type:boolean;default:false"`
}
