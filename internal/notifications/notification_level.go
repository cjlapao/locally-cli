package notifications

type NotificationLevel uint8

const (
	InfoLevel NotificationLevel = iota
	WarningLevel
	ErrorLevel
	CriticalLevel
	DebugLevel
	SuccessLevel
)

func (n NotificationLevel) String() string {
	switch n {
	case SuccessLevel:
		return "success"
	case InfoLevel:
		return "info"
	case WarningLevel:
		return "warn"
	case ErrorLevel:
		return "err"
	case CriticalLevel:
		return "crit"
	case DebugLevel:
		return "debug"
	default:
		return "unknown"
	}
}
