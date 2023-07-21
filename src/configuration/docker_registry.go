package configuration

type DockerRegistry struct {
	Enabled      bool                       `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Registry     string                     `json:"registry" yaml:"registry"`
	BasePath     string                     `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	ManifestPath string                     `json:"manifestPath,omitempty" yaml:"manifestPath,omitempty"`
	Tag          string                     `json:"-" yaml:"-"`
	Credentials  *DockerRegistryCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

func (s *DockerRegistry) Clone(source *DockerRegistry, all bool) {
	if s == nil {
		s = &DockerRegistry{}
	}

	if source == nil {
		if all {
			s = nil
		}
		return
	}
	s.Enabled = source.Enabled
	if source.Registry == "" {
		if all {
			s.Registry = source.Registry
		}
	} else {
		s.Registry = source.Registry
	}

	if source.ManifestPath == "" {
		if all {
			s.ManifestPath = source.ManifestPath
		}
	} else {
		s.ManifestPath = source.ManifestPath
	}

	if source.BasePath == "" {
		if all {
			s.BasePath = source.BasePath
		}
	} else {
		s.BasePath = source.BasePath
	}

	if source.Tag == "" {
		if all {
			s.Tag = source.Tag
		}
	} else {
		s.Tag = source.Tag
	}

	s.Credentials.Clone(source.Credentials, all)
}

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
