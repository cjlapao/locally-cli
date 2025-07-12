package context

import (
	"github.com/cjlapao/locally-cli/internal/notifications"
	"github.com/google/uuid"
)

var notify = notifications.Get()

func New(name string) *Context {
	result := Context{
		ID:        uuid.New().String(),
		IsValid:   false,
		IsEnabled: false,
		Name:      name,
	}

	return &result
}
