package azure_cli

type ResourceGroupResponse []ResourceGroupResponseElement

type ResourceGroupResponseElement struct {
	Name string `json:"name"`
}
