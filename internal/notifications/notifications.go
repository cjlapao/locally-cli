package notifications

import "time"

type Notification struct {
	Level     NotificationLevel
	State     NotificationState
	Message   string
	Timestamp time.Time
	Service   string
}

type Notifications struct {
	Items []Notification
}
