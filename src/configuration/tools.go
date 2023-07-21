package configuration

type Tools struct {
	Checked   *CheckedTools  `json:"-" yaml:"-"`
	Docker    *DockerTool    `json:"docker,omitempty" yaml:"docker,omitempty"`
	Caddy     *CaddyTool     `json:"caddy,omitempty" yaml:"caddy,omitempty"`
	Nuget     *NugetTool     `json:"nuget,omitempty" yaml:"nuget,omitempty"`
	Dotnet    *DotnetTool    `json:"dotnet,omitempty" yaml:"dotnet,omitempty"`
	Git       *GitTool       `json:"git,omitempty" yaml:"git,omitempty"`
	Terraform *TerraformTool `json:"terraform,omitempty" yaml:"terraform,omitempty"`
	AzureCli  *AzureCliTool  `json:"azurecli,omitempty" yaml:"azurecli,omitempty"`
	Npm       *NpmTool       `json:"npm,omitempty" yaml:"npm,omitempty"`
}

type CheckedTools struct {
	DockerChecked        bool
	DockerComposeChecked bool
	CaddyChecked         bool
	NugetChecked         bool
	DotnetChecked        bool
	GitChecked           bool
	TerraformChecked     bool
	AzureCliChecked      bool
	NpmChecked           bool
}

type DockerTool struct {
	BuildRetries int    `json:"buildRetries,omitempty" yaml:"buildRetries,omitempty"`
	DockerPath   string `json:"dockerPath,omitempty" yaml:"dockerPath,omitempty"`
	ComposerPath string `json:"dockerComposePath,omitempty" yaml:"dockerComposePath,omitempty"`
}

type CaddyTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type NugetTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type DotnetTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type GitTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type TerraformTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type AzureCliTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type NpmTool struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}
