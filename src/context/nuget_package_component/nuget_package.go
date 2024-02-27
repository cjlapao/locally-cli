package nuget_package_component

type NugetPackage struct {
	Source       string
	Name         string   `json:"name" yaml:"name"`
	MajorVersion string   `json:"majorVersion" yaml:"majorVersion"`
	ProjectFile  string   `json:"projectFile" yaml:"projectRoot"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}
