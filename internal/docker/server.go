package docker

type Server struct {
	Platform      Platform    `json:"Platform"`
	Components    []Component `json:"Components"`
	Version       string      `json:"Version"`
	APIVersion    string      `json:"ApiVersion"`
	MinAPIVersion string      `json:"MinAPIVersion"`
	GitCommit     string      `json:"GitCommit"`
	GoVersion     string      `json:"GoVersion"`
	OS            string      `json:"Os"`
	Arch          string      `json:"Arch"`
	KernelVersion string      `json:"KernelVersion"`
	BuildTime     string      `json:"BuildTime"`
}
