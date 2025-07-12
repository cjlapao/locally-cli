package types

import "time"

type User struct {
	BaseModel
	Name                  string    `gorm:"not null;type:text"`
	Username              string    `gorm:"not null;unique;type:text"`
	Password              string    `gorm:"not null;type:text"`
	Email                 string    `gorm:"not null;unique;type:text"`
	Role                  string    `gorm:"not null;type:text;default:'user'"`
	Status                string    `gorm:"not null;type:text;default:'active'"`
	TenantID              string    `gorm:"type:text"`
	TwoFactorEnabled      bool      `gorm:"type:boolean;not null;default:false"`
	TwoFactorSecret       string    `gorm:"type:text;not null;default:''"`
	TwoFactorVerified     bool      `gorm:"type:boolean;not null;default:false"`
	Blocked               bool      `gorm:"type:boolean;not null;default:false"`
	RefreshToken          string    `gorm:"type:text;not null;default:''"`
	RefreshTokenExpiresAt time.Time `gorm:"type:timestamp"`
}

func (User) TableName() string {
	return "users"
}
