package azure_cli

type WrapperLoggedInformation struct {
	LoggedIn           bool
	Username           string
	IsServicePrincipal bool
	SubscriptionId     string
	TenantId           string
}
