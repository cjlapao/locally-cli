// Package docker provides a service for interacting with the Docker daemon.
package docker

import (
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var (
	instance *DockerService
	once     sync.Once
)

type DockerService struct {
	sockSchema   string
	sockPath     string
	pollInterval time.Duration
	client       *client.Client
	monitor      *MonitorService
}

func Initialize() (*DockerService, error) {
	var initErr error
	once.Do(func() {
		svc, initErr := newService()
		if initErr == nil {
			instance = svc
		}
	})
	return instance, initErr
}

func GetInstance() *DockerService {
	if instance == nil {
		panic("docker service not initialized")
	}
	return instance
}

func newService() (*DockerService, error) {
	svc := &DockerService{
		sockSchema: "unix://",
		sockPath:   "/var/run/docker.sock",
	}
	dockerClient, err := client.NewClientWithOpts(client.WithHost(svc.sockSchema+svc.sockPath), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	svc.client = dockerClient
	svc.monitor, err = initializeMonitoringService(dockerClient)
	if err != nil {
		return nil, err
	}
	svc.monitor.Listen()
	return svc, nil
}

func (s *DockerService) WithDocker() *DockerService {
	s.sockPath = "/var/run/docker.sock"
	return s
}

func (s *DockerService) WithPodman() *DockerService {
	s.sockPath = "/run/podman/podman.sock"
	return s
}

func (s *DockerService) WithSocketPath(path string) *DockerService {
	s.sockPath = path
	return s
}

func (s *DockerService) ListContainers(ctx *appctx.AppContext) ([]container.Summary, error) {
	ctx.Log().Info("Listing containers")
	containers, err := s.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return []container.Summary{}, err
	}

	containerDTOs := make([]types.ContainerDTO, len(containers))
	for i, container := range containers {
		containerDTOs[i] = *containerToDTO(container)
	}

	return containers, nil
}

func (s *DockerService) CreateContainer(ctx *appctx.AppContext, name string, config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig) (*container.CreateResponse, error) {
	ctx.Log().Info("Creating container")
	container, err := s.client.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, name)
	if err != nil {
		return nil, err
	}
	return &container, nil
}

func (s *DockerService) StartContainer(ctx *appctx.AppContext, id string, options container.StartOptions) error {
	ctx.Log().WithField("container_id", id).Info("Starting container")
	err := s.client.ContainerStart(ctx, id, options)
	if err != nil {
		return err
	}
	return nil
}

func (s *DockerService) StopContainer(ctx *appctx.AppContext, id string, options container.StopOptions) error {
	ctx.Log().WithField("container_id", id).Info("Stopping container")
	err := s.client.ContainerStop(ctx, id, options)
	if err != nil {
		return err
	}
	return nil
}

func (s *DockerService) RemoveContainer(ctx *appctx.AppContext, id string, options container.RemoveOptions) error {
	ctx.Log().WithField("container_id", id).Info("Removing container from docker")
	err := s.client.ContainerRemove(ctx, id, options)
	if err != nil {
		return err
	}

	return nil
}
