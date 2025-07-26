package azure_cli

type StorageAccountResponse []ResourceGroupResponseElement

type StorageAccountResponseElement struct {
	Name string `json:"name"`
}
