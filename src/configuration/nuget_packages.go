package configuration

type NugetPackages struct {
	source       string
	OutputSource string          `json:"outputSource" yaml:"outputSource"`
	Packages     []*NugetPackage `json:"packages" yaml:"packages"`
}

type NugetPackage struct {
	source       string
	Name         string   `json:"name" yaml:"name"`
	MajorVersion string   `json:"majorVersion" yaml:"majorVersion"`
	ProjectFile  string   `json:"projectFile" yaml:"projectRoot"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}
