package docker

type Details struct {
	APIVersion    *string `json:"ApiVersion,omitempty"`
	Arch          *string `json:"Arch,omitempty"`
	BuildTime     *string `json:"BuildTime,omitempty"`
	Experimental  *string `json:"Experimental,omitempty"`
	GitCommit     string  `json:"GitCommit"`
	GoVersion     *string `json:"GoVersion,omitempty"`
	KernelVersion *string `json:"KernelVersion,omitempty"`
	MinAPIVersion *string `json:"MinAPIVersion,omitempty"`
	OS            *string `json:"Os,omitempty"`
}
