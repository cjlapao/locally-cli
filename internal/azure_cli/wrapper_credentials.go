package azure_cli

type WrapperCredentials struct {
	ServicePrincipal bool
	UseDeviceCode    bool
	Username         string
	Password         string
	SubscriptionId   string
	TenantId         string
}
