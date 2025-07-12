package docker

import (
	"context"
	"sync"

	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type MonitorService struct {
	client *client.Client
	stopCh chan struct{}
}

var (
	monitorInstance *MonitorService
	monitorOnce     sync.Once
)

func initializeMonitoringService(dockerClient *client.Client) (*MonitorService, error) {
	var initErr error
	monitorOnce.Do(func() {
		monitorInstance = newMonitoringService(dockerClient)
	})
	return monitorInstance, initErr
}

func getMonitoringInstance() *MonitorService {
	if monitorInstance == nil {
		panic("monitoring service not initialized")
	}
	return monitorInstance
}

func newMonitoringService(dockerClient *client.Client) *MonitorService {
	return &MonitorService{
		client: dockerClient,
		stopCh: make(chan struct{}),
	}
}

func (s *MonitorService) Listen() <-chan events.Message {
	logging.Info("Starting docker monitoring service")
	dockerEvents := make(chan events.Message, 32)
	go func() {
		defer close(dockerEvents)
		opts := events.ListOptions{
			Filters: filters.NewArgs(
				filters.Arg("type", "container"),
			),
		}
		eventsCh, errCh := s.client.Events(context.Background(), opts)
		for {
			select {
			case <-s.stopCh:
				return
			case event := <-eventsCh:
				if event.Type != events.ContainerEventType {
					continue
				}
				logging.WithFields(logrus.Fields{
					"event": event.Action,
				}).Debug("Received event")
				dockerEvents <- event
			case err := <-errCh:
				logging.WithError(err).Error("Error receiving events")
			}
		}
	}()

	return dockerEvents
}

func (s *MonitorService) Stop() error {
	close(s.stopCh)
	return nil
}
