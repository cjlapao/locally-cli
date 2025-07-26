package tools

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/cjlapao/locally-cli/internal/icons"

	"github.com/google/uuid"
)

type EmsApiKey struct {
	ApiKey   string    `json:"ApiKey"`
	TenantId uuid.UUID `json:"TenantId"`
}

type EmsServiceTool struct{}

var currentEmsServiceTool *EmsServiceTool

func GetEmsServiceTool() *EmsServiceTool {
	if currentEmsServiceTool == nil {
		currentEmsServiceTool = &EmsServiceTool{}
	}

	return currentEmsServiceTool
}

func (tool *EmsServiceTool) GenerateEmsApiKeyHeader(apiKey string) (string, error) {
	if apiKey == "" {
		err := errors.New("ApiKey cannot be empty, please use the flag --key to define it")
		logger.Error("%s %s", icons.IconRevolvingLight, err.Error())
		return "", err
	}

	tenantId, err := uuid.Parse("11112222-3333-4444-5555-666677778888")
	if err != nil {
		return "", err
	}

	emsApiKey := EmsApiKey{
		ApiKey:   apiKey,
		TenantId: tenantId,
	}

	jsonPayload, err := json.Marshal(emsApiKey)
	logger.Debug("%s EmsApiKey Payload: %s", icons.IconFire, string(jsonPayload))
	if err != nil {
		return "", err
	}

	generatedApiKey := base64.Encoding.WithPadding(*base64.StdEncoding, '=').EncodeToString(jsonPayload)
	return generatedApiKey, nil
}
