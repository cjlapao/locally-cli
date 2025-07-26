package azure_cli

type StorageAccountContainerResponse []ResourceGroupResponseElement

type StorageAccountContainerResponseElement struct {
	Name string `json:"name"`
}
