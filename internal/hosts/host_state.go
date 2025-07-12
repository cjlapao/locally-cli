package hosts

type HostState int

const (
	StateNew HostState = iota
	StateAdd
	StateNone
	StateClean
)
