package mock_component

type MockRouteResponse struct {
	ContentType string                 `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Body        map[string]interface{} `json:"body,omitempty" yaml:"body,omitempty"`
	RawBody     string                 `json:"rawBody,omitempty" yaml:"rawBody,omitempty"`
}
