package entities

type Role struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;type:text"`
	Description string `json:"description" gorm:"not null;type:text"`
	IsAdmin     bool   `json:"is_admin" gorm:"not null;type:boolean;default:false"`
	IsSuperUser bool   `json:"is_super_user" gorm:"not null;type:boolean;default:false"`
}

func (Role) TableName() string {
	return "roles"
}
