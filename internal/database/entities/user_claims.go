package entities

type UserClaims struct {
	BaseModel
	UserID  string `json:"user_id" gorm:"not null;type:text;index"`
	ClaimID string `json:"claim_id" gorm:"not null;type:text;index"`
}

func (UserClaims) TableName() string {
	return "user_claims"
}
