package notifications

type NotificationState uint8

const (
	NewState NotificationState = iota
	ReadState
)

func (n NotificationState) String() string {
	switch n {
	case NewState:
		return "new"
	case ReadState:
		return "read"
	default:
		return "unknown"
	}
}
