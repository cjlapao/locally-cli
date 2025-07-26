package docker_component

type DockerRegistryCredentials struct {
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	SubscriptionId string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantId       string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
}

func (s *DockerRegistryCredentials) Clone(source *DockerRegistryCredentials, all bool) {
	if s == nil {
		s = &DockerRegistryCredentials{}
	}

	if source == nil {
		if all {
			s = nil
		}
		return
	}

	if source.Username == "" {
		if all {
			s.Username = source.Username
		}
	} else {
		s.Username = source.Username
	}

	if source.Password == "" {
		if all {
			s.Password = source.Password
		}
	} else {
		s.Password = source.Password
	}

	if source.SubscriptionId == "" {
		if all {
			s.SubscriptionId = source.SubscriptionId
		}
	} else {
		s.SubscriptionId = source.SubscriptionId
	}

	if source.TenantId == "" {
		if all {
			s.TenantId = source.TenantId
		}
	} else {
		s.TenantId = source.TenantId
	}
}
