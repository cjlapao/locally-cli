package mappers

import (
	"fmt"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/notifications"
)

func DecodeGitCredentials(source *configuration.GitCredentials) *configuration.GitCredentials {
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

func DecodeBackendComponent(source *configuration.BackendComponent) *configuration.BackendComponent {
	env := environment.Get()
	notify := notifications.Get()
	source.Name = env.Replace(source.Name)

	notify.Debug("Decoded Backend Component : %v", fmt.Sprintf("%v", source))
	return source
}
