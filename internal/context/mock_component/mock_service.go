package mock_component

type MockService struct {
	Source     string
	Name       string      `json:"name,omitempty" yaml:"name,omitempty"`
	MockRoutes []MockRoute `json:"mockRoutes,omitempty" yaml:"mockRoutes,omitempty"`
}
