package docker

import (
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/docker/docker/api/types/container"
)

func containerToDTO(container container.Summary) *types.ContainerDTO {
	return &types.ContainerDTO{
		ID:    container.ID,
		Name:  container.Names[0],
		State: container.State,
	}
}
