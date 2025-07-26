package docker_component

type DockerCompose struct {
	Location string                           `json:"location,omitempty" yaml:"location,omitempty"`
	Version  string                           `json:"version,omitempty" yaml:"version,omitempty"`
	Name     string                           `json:"name,omitempty" yaml:"name,omitempty"`
	Services map[string]*DockerComposeService `json:"services,omitempty" yaml:"services,omitempty"`
}

func (s *DockerCompose) Clone(source *DockerCompose, all bool) {
	if s == nil {
		s = &DockerCompose{}
	}

	if source == nil {
		if all {
			s = nil
		}
		return
	}

	if source.Location == "" {
		if all {
			s.Location = source.Location
		}
	} else {
		s.Location = source.Location
	}

	if source.Version == "" {
		if all {
			s.Version = source.Version
		}
	} else {
		s.Version = source.Version
	}

	if source.Name == "" {
		if all {
			s.Name = source.Name
		}
	} else {
		s.Name = source.Name
	}

	for key, service := range s.Services {
		if s, ok := source.Services[key]; ok {
			service.Clone(s, all)
		}
	}
}
