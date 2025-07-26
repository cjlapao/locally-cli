package docker

type Client struct {
	Platform          Platform `json:"Platform"`
	CloudIntegration  string   `json:"CloudIntegration"`
	Version           string   `json:"Version"`
	APIVersion        string   `json:"ApiVersion"`
	DefaultAPIVersion string   `json:"DefaultAPIVersion"`
	GitCommit         string   `json:"GitCommit"`
	GoVersion         string   `json:"GoVersion"`
	OS                string   `json:"Os"`
	Arch              string   `json:"Arch"`
	BuildTime         string   `json:"BuildTime"`
	Context           string   `json:"Context"`
	Experimental      bool     `json:"Experimental"`
}
