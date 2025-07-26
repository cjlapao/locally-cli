package hosts

type CleanState int

const (
	STARTED CleanState = iota
	ENDED
	CLEANING
	NONE
)
