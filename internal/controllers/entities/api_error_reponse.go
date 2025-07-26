package entities

type ApiErrorResponse struct {
	Code       string `json:"code,omitempty" yaml:"code,omitempty"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
	Message    string `json:"message,omitempty" yaml:"message,omitempty"`
	StatusCode int    `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`
}

func NewApiErrorResponse(code string, message string, statusCode int) *ApiErrorResponse {
	result := ApiErrorResponse{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}

	return &result
}

func NewApiErrorResponseFromError(code string, err error) *ApiErrorResponse {
	result := ApiErrorResponse{
		Error:      "system_exception",
		Code:       code,
		Message:    err.Error(),
		StatusCode: 500,
	}

	return &result
}
