package caddy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/helpers"

	"github.com/cjlapao/common-go/helper"
)

type CaddyCommandWrapper struct {
	ToolPath string
	Output   string
}

func GetWrapper() *CaddyCommandWrapper {
	config = configuration.Get()
	return &CaddyCommandWrapper{}
}

func (svc *CaddyCommandWrapper) Run() error {
	notify.Rocket("Running Caddy...")
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))
	for _, client := range config.GetCurrentContext().SpaServices {
		notify.Debug("Setting env variables for client %v=%v", fmt.Sprintf("%v_path", common.EncodeName(client.Name)), client.Path)
		os.Setenv(fmt.Sprintf("%v_path", common.EncodeName(client.Name)), client.Path)
	}

	output, err := executer.ExecuteAndWatch(helpers.GetCaddyPath(), "run", "--config", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, "Caddyfile"))

	if err != nil {
		notify.FromError(err, "Something wrong running caddy")
		return err
	}

	svc.Output = output.GetAllOutput()

	return nil
}

func (svc *CaddyCommandWrapper) Stop() error {
	notify.Info("Closing caddy")
	client := http.Client{}
	baseUrl, _ := url.Parse("http://127.0.0.1/stop")
	request := http.Request{
		Method: "POST",
		URL:    baseUrl,
	}

	_, err := client.Do(&request)

	notify.Info("Caddy Closed")
	return err
}
