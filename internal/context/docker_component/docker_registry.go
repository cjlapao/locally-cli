package docker_component

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
