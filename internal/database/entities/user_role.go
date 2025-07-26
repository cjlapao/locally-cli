package entities

type UserRole struct {
	BaseModel
	UserID string `json:"user_id" gorm:"not null;type:text;index"`
	RoleID string `json:"role_id" gorm:"not null;type:text;index"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
