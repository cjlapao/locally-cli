package operations

import (
	"strings"

	"github.com/cjlapao/locally-cli/internal/interfaces"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

var notify = notifications.Get()

var globalOperations *OperationService

type OperationService struct {
	operations []interfaces.Operation
}

func Get() *OperationService {
	if globalOperations == nil {
		globalOperations = &OperationService{
			operations: make([]interfaces.Operation, 0),
		}

		globalOperations.Register(NewApiOperation())
	}

	return globalOperations
}

func (service *OperationService) Register(operation interfaces.Operation) {
	service.operations = append(service.operations, operation)
}

func (service *OperationService) GetOperation(name string) interfaces.Operation {
	for _, operation := range service.operations {
		if strings.EqualFold(operation.GetName(), name) {
			return operation
		}
	}

	return nil
}
