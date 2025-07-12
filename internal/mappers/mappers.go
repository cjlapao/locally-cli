package mappers

import (
	"fmt"

	"github.com/cjlapao/locally-cli/internal/context/git_component"
	"github.com/cjlapao/locally-cli/internal/context/service_component"
	"github.com/cjlapao/locally-cli/internal/environment"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

func DecodeGitCredentials(source *git_component.GitCredentials) *git_component.GitCredentials {
	env := environment.Get()
	notify := notifications.Get()
	source.AccessToken = env.Replace(source.AccessToken)
	source.Password = env.Replace(source.Password)
	source.Username = env.Replace(source.Username)
	source.PublicKeyPath = env.Replace(source.PublicKeyPath)
	source.PrivateKeyPath = env.Replace(source.PrivateKeyPath)

	notify.Debug("Decoded Git Credentials: %v", fmt.Sprintf("%v", source))
	return source
}

func DecodeBackendComponent(source *service_component.BackendComponent) *service_component.BackendComponent {
	env := environment.Get()
	notify := notifications.Get()
	source.Name = env.Replace(source.Name)

	notify.Debug("Decoded Backend Component : %v", fmt.Sprintf("%v", source))
	return source
}
