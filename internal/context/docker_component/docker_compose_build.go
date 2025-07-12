package docker_component

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
