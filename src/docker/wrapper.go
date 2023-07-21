package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/azure_cli"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/icons"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

type DockerCommandWrapper struct {
	Output string
}

func GetWrapper() *DockerCommandWrapper {
	svc := &DockerCommandWrapper{}

	return svc
}

func (svc *DockerCommandWrapper) Build(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output string
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Build for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Build for service %v on path %v", serviceName, path)
	}

	if helper.GetFlagSwitch("force-clean", false) {
		images, err := svc.GetServiceImages(path, serviceName, componentName)
		if err != nil {
			return err
		}

		for _, image := range images {
			if err := svc.RemoveImage(image.Repository, image.Tag); err != nil {
				return err
			}
		}
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "build")
	if componentName != "" {
		args = append(args, componentName)
	}
	if helper.GetFlagSwitch("no-cache", false) {
		notify.Info("Using no cache to build images")
		args = append(args, "--no-cache")
	}

	output, err = configuration.Retry("Docker Compose Build", configuration.GetDockerComposePath(), args, true)

	if err != nil {
		notify.FromError(err, "There was an error running docker build")
		return err
	}

	svc.Output = output

	return err
}

func (svc *DockerCommandWrapper) Up(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output string
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Up for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Up for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "up")
	if componentName != "" {
		args = append(args, componentName)
	}
	args = append(args, "-d")
	if helper.GetFlagSwitch("force-recreate", false) {
		notify.InfoWithIcon(icons.IconFlag, "Forcing recreate of containers")
		args = append(args, "--force-recreate")
	}

	output, err = configuration.Retry("docker-compose Up", configuration.GetDockerComposePath(), args, true)

	if err != nil {
		notify.FromError(err, "There was an error running docker-compose up")
		return err
	}

	svc.Output = output

	return err
}

func (svc *DockerCommandWrapper) Down(path string, serviceName string, componentName string) error {
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Down for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Down for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "down")

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose build")
		return err
	}

	svc.Output = output.GetAllOutput()

	return err
}

func (svc *DockerCommandWrapper) Start(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Start for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Start for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "start")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose start")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Stop(path string, serviceName string, componentName string) error {
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Info("Running Docker Compose Stop for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Info("Running Docker Compose Stop for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "stop")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose stop")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Pause(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Pause for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Pause for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "pause")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose pause")
		return err
	}

	svc.Output = output.GetAllOutput()

	return err
}

func (svc *DockerCommandWrapper) Resume(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Unpause for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Unpause for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "unpause")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose unpause")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Status(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose Status for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Status for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "ps")
	args = append(args, "--format")
	args = append(args, "pretty")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker compose component status")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) IsRunning(imageName string) (bool, error) {
	env := environment.Get()
	var output executer.ExecuteOutput
	var err error

	imageName = env.Replace(imageName)

	if imageName == "" {
		return false, errors.New("path cannot be empty or nil")
	}

	args := make([]string, 0)
	args = append(args, "ps")
	args = append(args, "--all")
	args = append(args, "--no-trunc")
	args = append(args, "--format")
	args = append(args, "'{\"image\": \"{{.Image}}\", \"state\":\"{{.State}}\"}'")

	output, err = executer.ExecuteWithNoOutput(configuration.GetDockerPath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker compose component status")
		return false, err
	}

	lines := strings.Split(output.StdOut, "\n")
	for _, line := range lines {
		line := strings.Trim(line, "'")
		if config.Debug() {
			notify.Debug(line)
		}

		if line == "" {
			continue
		}

		var service RunningResponse
		if err := json.Unmarshal([]byte(line), &service); err != nil {
			return false, err
		}
		if strings.EqualFold(imageName, service.Image) && strings.EqualFold("running", service.State) {
			return true, nil
		}
	}

	svc.Output = output.StdOut

	return false, err
}

func (svc *DockerCommandWrapper) List(serviceName string) error {
	env := environment.Get()

	serviceName = env.Replace(serviceName)
	var output executer.ExecuteOutput
	var err error

	if serviceName != "" {
		notify.Rocket("Running Docker Compose Status for service %v ", serviceName)
	} else {
		notify.Rocket("Running Docker Compose Status for all services")
	}

	args := make([]string, 0)
	args = append(args, "ls")
	args = append(args, "--format")
	args = append(args, "pretty")
	if serviceName != "" {
		args = append(args, "--filter")
		args = append(args, fmt.Sprintf("name=%v", serviceName))
	}
	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker compose component status")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Logs(path string, serviceName string, componentName string) error {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)

	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return errors.New("path cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose logs for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose Up for service %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "logs")
	if componentName != "" {
		args = append(args, componentName)
	}
	if helper.GetFlagSwitch("follow", false) {
		notify.Info("Following the logs, use CTRL + C to stop")
		args = append(args, "-f")
	}

	output, err = executer.Execute(configuration.GetDockerComposePath(), args...)

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose logs")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) GetServiceImages(path string, serviceName string, componentName string) ([]ContainerImage, error) {
	env := environment.Get()

	path = env.Replace(path)
	serviceName = env.Replace(serviceName)
	componentName = env.Replace(componentName)
	var output executer.ExecuteOutput
	var err error

	if path == "" {
		return nil, errors.New("path cannot be empty or nil")
	}

	if serviceName == "" {
		return nil, errors.New("service name cannot be empty or nil")
	}

	if componentName != "" {
		notify.Rocket("Running Docker Compose image list for component %v of service %v on path %v", componentName, serviceName, path)
	} else {
		notify.Rocket("Running Docker Compose image list %v on path %v", serviceName, path)
	}

	args := make([]string, 0)
	args = append(args, "--project-directory")
	args = append(args, path)
	args = append(args, "images")
	if componentName != "" {
		args = append(args, componentName)
	}

	output, err = executer.ExecuteWithNoOutput(configuration.GetDockerComposePath(), args...)

	var parsedOutput = make([]ContainerImage, 0)
	lines := strings.Split(output.StdOut, "\n")
	for lineNumb, line := range lines {
		lineParts := strings.Split(line, " ")
		containerImage := ContainerImage{}
		idx := 0
		if lineNumb == 0 {
			continue
		}
		for _, linePart := range lineParts {
			linePart = strings.TrimSpace(linePart)
			if linePart != "" {
				switch idx {
				case 0:
					containerImage.Container = linePart
					idx += 1
				case 1:
					containerImage.Repository = linePart
					idx += 1
				case 2:
					containerImage.Tag = linePart
					idx += 1
				case 3:
					containerImage.ImageId = linePart
					idx += 1
				case 4:
					containerImage.Size = linePart
					idx += 1
				}
			}
		}

		if containerImage.Container != "" {
			parsedOutput = append(parsedOutput, containerImage)
		}

	}

	if err != nil {
		notify.FromError(err, "Something wrong running docker compose logs")
		return nil, err
	}

	svc.Output = output.StdOut

	return parsedOutput, err
}

func (svc *DockerCommandWrapper) RemoveImage(imageName string, tagName string) error {
	env := environment.Get()
	var output executer.ExecuteOutput
	var err error

	if imageName == "" {
		err = errors.New("image name cannot be empty")
		notify.Error(err.Error())
		return err
	}

	imageName = env.Replace(imageName)
	tagName = env.Replace(tagName)
	notify.InfoWithIcon(icons.IconBomb, "Removing image %v:%v from system repository", imageName, tagName)
	if tagName == "" {
		tagName = "latest"
	}

	args := make([]string, 0)
	args = append(args, "image")
	args = append(args, "rm")
	args = append(args, fmt.Sprintf("%v:%v", imageName, tagName))

	output, err = executer.Execute(configuration.GetDockerPath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker image removal")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) BuildImage(options BuildImageOptions) error {
	var output executer.ExecuteOutput
	var err error

	args, err := options.GetArguments()
	if err != nil {
		return err
	}

	notify.Rocket("Building image %v:%v", options.Name, options.Tag)

	notify.Debug("Parameters: %s", fmt.Sprintf("%v", args))
	output, err = executer.Execute(configuration.GetDockerPath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker build image")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Login(crName, username, password, subscriptionId, tenantId string) error {
	env := environment.Get()
	var output executer.ExecuteOutput
	var err error

	if crName == "" {
		err = errors.New("container registry cannot be empty")
		notify.Error(err.Error())
		return err
	}

	if username == "" {
		err = errors.New("user name cannot be empty")
		notify.Error(err.Error())
		return err
	}

	if password == "" {
		err = errors.New("password cannot be empty")
		notify.Error(err.Error())
		return err
	}

	notify.Debug("username %s, password, %s", username, password)
	crName = env.Replace(crName)
	username = env.Replace(username)
	password = env.Replace(password)

	// Processing password to check if we need to exchange it for a token
	password, err = processAzureAcrOauthPassword(password, subscriptionId, tenantId)
	if err != nil {
		return err
	}

	notify.Rocket("Logging in to container registry %s with user %s", crName, username)

	args := make([]string, 0)
	args = append(args, "login")
	args = append(args, "-u")
	args = append(args, username)
	args = append(args, "-p")
	args = append(args, password)
	args = append(args, crName)

	output, err = executer.Execute(configuration.GetDockerPath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker login")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func (svc *DockerCommandWrapper) Pull(crName, imagePath, tag string) error {
	env := environment.Get()

	crName = env.Replace(crName)
	imagePath = env.Replace(imagePath)
	tag = env.Replace(tag)

	var output executer.ExecuteOutput
	var err error

	if crName == "" {
		err = errors.New("container registry cannot be empty")
		notify.Error(err.Error())
		return err
	}

	crName = strings.ReplaceAll(strings.ReplaceAll(crName, "https://", ""), "http://", "")
	if imagePath == "" {
		err = errors.New("image path cannot be empty")
		notify.Error(err.Error())
		return err
	}

	if tag == "" {
		err = errors.New("tag cannot be empty")
		notify.Error(err.Error())
		return err
	}

	notify.Rocket("Pulling image %s:%s from container registry %s", imagePath, tag, crName)

	args := make([]string, 0)
	args = append(args, "pull")
	crName = strings.Trim(crName, "/")
	imagePath = strings.Trim(imagePath, "/")

	args = append(args, fmt.Sprintf("%s/%s:%s", crName, imagePath, tag))

	notify.Debug("Pull arguments: %s", fmt.Sprintf("%v", args))
	output, err = executer.Execute(configuration.GetDockerPath(), args...)
	if err != nil {
		notify.FromError(err, "Something wrong running docker pull")
		return err
	}

	svc.Output = output.StdOut

	return err
}

func processAzureAcrOauthPassword(password, subscriptionId, tenantId string) (string, error) {
	azureCli := azure_cli.Get()
	// Checking for special password
	if strings.HasPrefix(password, environment.PREFIX) && strings.HasSuffix(password, environment.SUFFIX) {
		password = strings.TrimPrefix(password, environment.PREFIX)
		password = strings.TrimSuffix(password, environment.SUFFIX)
		password = strings.TrimSpace(password)
		parts := strings.Split(password, ".")
		if strings.EqualFold(parts[0], "azure") {
			key := ""
			for i := 1; i < len(parts); i++ {
				if i > 1 {
					key += "."
				}
				key += parts[i]
			}
			notify.Debug("Found environment key: %s", key)
			if strings.HasPrefix(key, "acr.") && strings.HasSuffix(key, ".token") {
				notify.Debug("Matched against special key for acr token")
				acrName := strings.TrimPrefix(key, "acr.")
				acrName = strings.TrimSuffix(acrName, ".token")
				notify.Debug("ACR name: %s", acrName)
				token, err := azureCli.GetAcrRefreshToken(acrName, subscriptionId, tenantId)
				if err != nil {
					return "", err
				}

				password = token
				notify.Debug("Password replaced with token: %s", password)
			}
		}
	}

	return password, nil
}
