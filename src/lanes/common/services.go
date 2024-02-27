package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/environment"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func RequestOpsToken(baseurl string) (string, error) {
	env := environment.Get()

	var ops_request *http.Request
	_opsurl := baseurl + "/ops/connect/token"

	_form := url.Values{}
	_form.Set("client_id", env.Replace("${{ keyvault.global--environment--ops--url }}"))
	_form.Set("client_secret", env.Replace("${{ keyvault.global--environment--ops--password }}"))
	_form.Set("grant_type", "client_credentials")
	_body_data := strings.NewReader(_form.Encode())

	ops_request, err := http.NewRequest("POST", _opsurl, _body_data)
	if err != nil {
		return "", err
	}

	ops_request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	ops_client := &http.Client{}
	ops_response, err := ops_client.Do(ops_request)
	if err != nil {
		return "", err
	}
	defer ops_response.Body.Close()

	if ops_response.StatusCode != 200 {
		return "", errors.New("Could not retrieve Ops token")
	}

	ops_response_body, err := ioutil.ReadAll(ops_response.Body)
	if err != nil {
		return "", err
	}

	var ops_response_json OpsResponse
	if err := json.Unmarshal(ops_response_body, &ops_response_json); err != nil {
		return "", err
	}

	access_token := ops_response_json.AccessToken

	return access_token, nil
}

func SendPostRequest(access_token string, content []byte, url string) (string, error) {
	var request *http.Request

	request, err := http.NewRequest("POST", url, strings.NewReader(string(content)))
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", access_token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	status_code := fmt.Sprintf("%d", response.StatusCode)

	return status_code, nil
}
