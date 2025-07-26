package common

import "github.com/cjlapao/common-go/helper"

const (
	DEFAULT_locally_OUTPUT_PATH     string = ".cache-data"
	DEFAULT_CONFIG_SERVICE_PATH     string = "config-data"
	CADDY_PATH                      string = "caddy"
	INFRASTRUCTURE_PATH             string = "infrastructure"
	PIPELINES_PATH                  string = "pipelines"
	CADDY_UI_PATH                   string = "webclients"
	CADDY_ROOT_SERVICES_PATH        string = "root_services"
	CADDY_ROOT_SERVICES_HOSTS_PATH  string = "hosts"
	CADDY_ROOT_SERVICES_ROUTES_PATH string = "routes"
	CADDY_MOCK_ROUTES_PATH          string = "mocked_routes"
	CADDY_TENANTS_PATH              string = "tenants"
	CADDY_HOSTED_SERVICES_PATH      string = "hosted_services"
	SPA_PATH                        string = "webclients"
	TLS_PATH                        string = "ssl"
	WEB_CLIENT_SHELL_NAME           string = "WebClient Shell"
	SOURCES_PATH                    string = "sources"
	DOCKER_COMPOSE_PATH             string = "docker_compose"
	SERVICE_NAME                    string = "locally"
	DEFAULT_RETRY_COUNT             int    = 3
	DEFAULT_WAITING_FOR_SECONDS     int    = 5
	OUTPUT_TO_FILE                  string = "outputToFile"
)

const (
	API_PREFIX_VAR string = "API_PREFIX"
	API_PORT_VAR   string = "API_PORT"
)

const (
	OVERRIDE_CONFIG_FILE_MARKER string = ".override"
)

func IsVerbose() bool {
	return helper.GetFlagSwitch("verbose", false)
}

func IsDebug() bool {
	return helper.GetFlagSwitch("debug", false)
}
