package hosts

type HostEntry struct {
	IP        string
	Hosts     []string
	Comment   string
	IsNew     bool
	InSection bool
	State     HostState
}
