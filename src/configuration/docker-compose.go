package configuration

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

type DockerComposeService struct {
	Image       string                 `json:"image,omitempty" yaml:"image,omitempty"`
	Build       *DockerComposeBuild    `json:"build,omitempty" yaml:"build,omitempty"`
	Volumes     []string               `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Ports       []string               `json:"ports,omitempty" yaml:"ports,omitempty"`
	Environment map[string]interface{} `json:"environment,omitempty" yaml:"environment,omitempty"`
}

func (s *DockerComposeService) Clone(source *DockerComposeService, all bool) {
	if s == nil {
		s = &DockerComposeService{}
	}

	if source == nil {
		if all {
			s = nil
		}
		return
	}

	if source.Image == "" {
		if all {
			s.Image = source.Image
		}
	} else {
		s.Image = source.Image
	}
	s.Build.Clone(source.Build, all)

	if len(source.Volumes) == 0 {
		if all {
			s.Volumes = source.Volumes
		}
	} else {
		s.Volumes = source.Volumes
	}

	if len(source.Ports) == 0 {
		if all {
			s.Ports = source.Ports
		}
	} else {
		s.Ports = source.Ports
	}

	if len(source.Environment) == 0 {
		if all {
			s.Environment = source.Environment
		}
	} else {
		s.Environment = source.Environment
	}
}

type DockerComposeBuild struct {
	Context    string   `json:"context,omitempty" yaml:"context,omitempty"`
	Dockerfile string   `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	Args       []string `json:"args,omitempty" yaml:"args,omitempty"`
}

func (s *DockerComposeBuild) Clone(source *DockerComposeBuild, all bool) {
	if s == nil {
		s = &DockerComposeBuild{}
	}

	if source == nil {
		if all {
			s = nil
		}
		return
	}
	if source.Context == "" {
		if all {
			s.Context = source.Context
		}
	} else {
		s.Context = source.Context
	}

	if source.Dockerfile == "" {
		if all {
			s.Dockerfile = source.Dockerfile
		}
	} else {
		s.Dockerfile = source.Dockerfile
	}

	if len(source.Args) == 0 {
		if all {
			s.Args = source.Args
		}
	} else {
		s.Args = source.Args
	}
}
