package tenant

type TenantUpdateRequest struct {
	ID           string `json:"id" yaml:"id"`
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description" yaml:"description"`
	Domain       string `json:"domain" yaml:"domain"`
	OwnerID      string `json:"owner_id" yaml:"owner_id"`
	ContactEmail string `json:"contact_email" yaml:"contact_email"`
}
