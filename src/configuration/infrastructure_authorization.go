package configuration

type InfrastructureAuthorization struct {
	ClientId       string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret   string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	SubscriptionId string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantId       string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
}
