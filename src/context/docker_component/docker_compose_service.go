package docker_component

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
