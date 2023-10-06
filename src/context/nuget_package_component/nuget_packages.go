package nuget_package_component

type NugetPackages struct {
	Source       string
	OutputSource string          `json:"outputSource" yaml:"outputSource"`
	Packages     []*NugetPackage `json:"packages" yaml:"packages"`
}
