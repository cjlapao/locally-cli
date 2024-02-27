package azure_service

import (
	"context"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/cjlapao/locally-cli/entities"
)

var globalInstance *AzureService
var mu sync.Mutex

type AzureService struct {
}

func New() *AzureService {
	result := AzureService{}

	return &result
}

func Get() *AzureService {
	mu.Lock()
	defer mu.Unlock()

	if globalInstance == nil {
		globalInstance = New()
	}
	return globalInstance
}

func (c *AzureService) Name() string {
	return "azure"
}

func (c *AzureService) GetCredentials(credentials entities.AzureCredentials) (*azidentity.ClientSecretCredential, error) {
	credential, err := azidentity.NewClientSecretCredential(
		credentials.TenantId,
		credentials.ClientId,
		credentials.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{},
	)

	if err != nil {
		return nil, err
	}

	return credential, nil
}

func (c *AzureService) TestConnection(credentials entities.AzureCredentials) error {
	azureCredentials, err := c.GetCredentials(credentials)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)

	defer cancel()

	_, err = azureCredentials.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})

	if err != nil {
		return err
	}

	return nil
}
