package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/executer"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
)

type AzureAcrTool struct {
}

type AcrLoginResponse struct {
	AccessToken string `json:"accessToken"`
	LoginServer string `json:"loginServer"`
}

type Oauth2ACRTokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
}

func (svc *AzureAcrTool) GetAcrRefreshToken(acr, subscription string) (string, error) {
	config := configuration.Get()

	notify.Rocket("Running Azure Cli Acr Token...")

	if acr == "" {
		err := errors.New("acr cannot be null or empty")
		return "", err
	}

	acr = strings.TrimPrefix(acr, "https://")
	acr = strings.TrimPrefix(acr, "http://")

	encodedAcrName := configuration.EncodeName(acr)
	token := os.Getenv(fmt.Sprintf("locally_AZURE_%s_ACR_TOKEN", encodedAcrName))
	notify.Debug("Token value: %s", token)

	if token != "" {
		tokenBytes := []byte(token)
		rawToken, _ := jwt.ParseWithoutCheck(tokenBytes)
		notify.Debug("Raw token: %v", rawToken)
		minus1Minute := time.Now().Add((time.Minute * 1) * -1)
		isExpired := rawToken.Expires.Time().Before(minus1Minute)
		if !isExpired {
			notify.Debug("Using same token as it was not expired")
			return token, nil
		}
	}

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "acr")
	runArgs = append(runArgs, "login")
	runArgs = append(runArgs, "-n")
	runArgs = append(runArgs, acr)
	if subscription != "" {
		runArgs = append(runArgs, "--subscription")
		runArgs = append(runArgs, subscription)
	}
	runArgs = append(runArgs, "--expose-token")

	if config.Debug() {
		notify.Debug("run parameters: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteWithNoOutput(configuration.GetAzureCliPath(), runArgs...)

	if err != nil {
		notify.FromError(err, "Something wrong running setting the subscription")
		if output.GetAllOutput() != "" {
			notify.Error(output.GetAllOutput())
		}
		return "", err
	}

	var response AcrLoginResponse
	if err := json.Unmarshal([]byte(output.StdOut), &response); err != nil {
		notify.Error("failed to unmarshal the response")
	}

	os.Setenv(fmt.Sprintf("locally_AZURE_%s_ACR_TOKEN", encodedAcrName), response.AccessToken)

	return response.AccessToken, nil
}

func (svc *AzureAcrTool) ExchangeRefreshTokenForAccessToken(acr, scope string) (string, error) {
	notify.Rocket("Running Azure Refresh token exchange")

	if acr == "" {
		err := errors.New("subscription id cannot be null or empty")
		return "", err
	}

	token, _ := svc.GetAcrRefreshToken(acr, "")

	notify.Debug(token)
	acr = strings.TrimPrefix(acr, "http://")
	acr = strings.TrimPrefix(acr, "https://")

	if !strings.Contains(acr, ".azurecr.io") {
		acr = fmt.Sprintf("%s.azurecr.io", acr)
		notify.Debug("not found the domain, added it", acr)
	}

	if scope == "" {
		scope = "repository:*:metadata_read"
	}
	oauth2Endpoint := fmt.Sprintf("https://%s/oauth2/token", acr)
	body := url.Values{}
	body.Add("grant_type", "refresh_token")
	body.Add("service", acr)
	body.Add("scope", scope)
	body.Add("refresh_token", token)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	defer cancel()

	notify.Debug("Body: %v", body.Encode())
	notify.Debug("Host: %v", oauth2Endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", oauth2Endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid http response, got %s", fmt.Sprintf("%v", resp.StatusCode))
	}

	if resp.Body == nil {
		return "", fmt.Errorf("body cannot be nil")
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	notify.Debug("Parsing the response body %s", fmt.Sprintf("%v", string(respBody)))
	var response Oauth2ACRTokenExchangeResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", err
	}

	return response.AccessToken, nil
}
