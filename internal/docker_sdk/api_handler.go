package docker

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
)

type APIHandler struct {
	service *DockerService
}

func NewApiHandler(service *DockerService) *APIHandler {
	return &APIHandler{service: service}
}

func (h *APIHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/docker/containers",
			Handler:     h.GetAllContainers,
			Description: "Get all containers",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/docker/containers",
			Handler:     h.CreateContainer,
			Description: "Create a new container",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/docker/containers/{id}/start",
			Handler:     h.StartContainer,
			Description: "Start a container",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/docker/containers/{id}/stop",
			Handler:     h.StopContainer,
			Description: "Stop a container",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/docker/containers/{id}",
			Handler:     h.RemoveContainer,
			Description: "Remove a container",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
	}
}

func (h *APIHandler) GetAllContainers(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	containers, err := h.service.ListContainers(ctx)
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to list containers", err.Error())
		return
	}

	containerDTOs := make([]types.ContainerDTO, len(containers))
	for i, container := range containers {
		containerDTOs[i] = *containerToDTO(container)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(containerDTOs)
}

func (h *APIHandler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	var request types.ContainerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteBadRequest(w, r, "Failed to decode request", err.Error())
		return
	}

	if errors := validation.Validate(request); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	createdContainer, err := h.service.CreateContainer(ctx, request.Name, &container.Config{
		Image: request.Image,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%s/%s", request.ExposePorts[0].ContainerPort, "tcp")): []nat.PortBinding{
				{
					HostPort: request.ExposePorts[0].HostPort,
					HostIP:   "0.0.0.0",
				},
			},
		},
	}, nil)
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to create container", err.Error())
		return
	}

	if request.RunOnCreate {
		err = h.service.StartContainer(ctx, createdContainer.ID, container.StartOptions{})
		if err != nil {
			api.WriteBadRequest(w, r, "Failed to start container", err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdContainer)
}

func (h *APIHandler) StartContainer(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	containerID := mux.Vars(r)["id"]
	err := h.service.StartContainer(ctx, containerID, container.StartOptions{})
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to start container", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.ContainerOperationResponse{
		ContainerID: containerID,
		Success:     true,
	})
}

func (h *APIHandler) StopContainer(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	containerID := mux.Vars(r)["id"]

	err := h.service.StopContainer(ctx, containerID, container.StopOptions{})
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to stop container", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.ContainerOperationResponse{
		ContainerID: containerID,
		Success:     true,
	})
}

func (h *APIHandler) RemoveContainer(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	var request types.ContainerRemoveRequest
	containerID := mux.Vars(r)["id"]

	if containerID == "" {
		api.WriteBadRequest(w, r, "Container ID is required", "")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteBadRequest(w, r, "Failed to decode request", err.Error())
		return
	}

	if errors := validation.Validate(request); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	err := h.service.RemoveContainer(ctx, containerID, container.RemoveOptions{
		Force:         request.Force,
		RemoveVolumes: request.RemoveVolumes,
		RemoveLinks:   request.RemoveLinks,
	})
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to remove container", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.ContainerOperationResponse{
		ContainerID: containerID,
		Success:     true,
	})
}
