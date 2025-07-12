package azure_cli

type AzureLoginResponse []AzureLoginResponseElement

type AzureLoginResponseElement struct {
	CloudName        CloudName         `json:"cloudName"`
	HomeTenantID     string            `json:"homeTenantId"`
	ID               string            `json:"id"`
	IsDefault        bool              `json:"isDefault"`
	ManagedByTenants []ManagedByTenant `json:"managedByTenants"`
	Name             string            `json:"name"`
	State            State             `json:"state"`
	TenantID         string            `json:"tenantId"`
	User             UserClass         `json:"user"`
}

type ManagedByTenant struct {
	TenantID string `json:"tenantId"`
}

type UserClass struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CloudName string

const (
	AzureCloud CloudName = "AzureCloud"
)

type State string

const (
	Enabled State = "Enabled"
)
