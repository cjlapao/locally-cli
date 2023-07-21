package notifications

import (
	"fmt"
	"github.com/cjlapao/locally-cli/icons"
	"strings"
	"time"
)

type NotificationsService struct {
	Service       string
	notifications *Notifications
}

func New(service string) *NotificationsService {
	if globalNotifications == nil {
		globalNotifications = &Notifications{
			Items: make([]Notification, 0),
		}
	}

	svc := NotificationsService{
		Service:       service,
		notifications: globalNotifications,
	}

	return &svc
}

func Get() *NotificationsService {
	return New("System")
}

func (svc *NotificationsService) Reset() {
	for _, n := range svc.notifications.Items {
		n.State = ReadState
	}
}

func (svc *NotificationsService) InfoWithIcon(icon, message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icon, "", words...)
}

func (svc *NotificationsService) InfoIndent(message, indentation string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconBlackSquare, indentation, words...)
}
func (svc *NotificationsService) InfoIndentIcon(icon, message, indentation string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icon, indentation, words...)
}

func (svc *NotificationsService) Info(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconBlackSquare, "", words...)
}

func (svc *NotificationsService) Wrench(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconWrench, "", words...)
}

func (svc *NotificationsService) Hammer(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconHammer, "", words...)
}

func (svc *NotificationsService) Rocket(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconRocket, "", words...)
}

func (svc *NotificationsService) Flag(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconFlag, "", words...)
}

func (svc *NotificationsService) MagnifyingGlass(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconMagnifyingGlass, "", words...)
}

func (svc *NotificationsService) Success(message string, words ...interface{}) {
	svc.Notify(true, SuccessLevel, message, icons.IconThumbsUp, "", words...)
}

func (svc *NotificationsService) Failure(message string, words ...interface{}) {
	svc.Notify(true, InfoLevel, message, icons.IconThumbDown, "", words...)
}

func (svc *NotificationsService) Warning(message string, words ...interface{}) {
	svc.Notify(true, WarningLevel, message, icons.IconWarning, "", words...)
}

func (svc *NotificationsService) Error(message string, words ...interface{}) {
	svc.Notify(true, ErrorLevel, message, icons.IconRevolvingLight, "", words...)
}

func (svc *NotificationsService) AddError(message string, words ...interface{}) {
	svc.Notify(true, ErrorLevel, message, icons.IconRevolvingLight, "", words...)
}

func (svc *NotificationsService) FromError(err error, message string, words ...interface{}) {
	msg := fmt.Sprintf("%s, err. %s", message, err.Error())
	svc.Notify(true, ErrorLevel, msg, icons.IconRevolvingLight, "", words...)
}

func (svc *NotificationsService) Critical(message string, words ...interface{}) {
	svc.Notify(true, CriticalLevel, message, icons.IconRevolvingLight, "", words...)
}

func (svc *NotificationsService) Debug(message string, words ...interface{}) {
	svc.Notify(true, DebugLevel, message, icons.IconFire, "", words...)
}

func (svc *NotificationsService) Notify(print bool, level NotificationLevel, message, icon, indentation string, words ...interface{}) {
	if len(words) > 0 {
		message = fmt.Sprintf(message, words...)
	}

	found := false
	for _, n := range svc.notifications.Items {
		if strings.EqualFold(level.String(), n.Level.String()) && strings.EqualFold(message, n.Message) && strings.EqualFold(svc.Service, n.Service) {
			found = true
			break
		}
	}

	if !found {
		svc.notifications.Items = append(svc.notifications.Items, Notification{
			Level:     level,
			Message:   message,
			Timestamp: time.Now(),
			Service:   svc.Service,
		})
	}

	if print {
		switch level {
		case InfoLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Info("%s", message)
			} else {
				logger.Info("%s %s", icon, message)
			}
		case WarningLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Warn("%s", message)
			} else {
				logger.Warn("%s %s", icon, message)
			}
		case ErrorLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Error("%s", message)
			} else {
				logger.Error("%s %s", icon, message)
			}
		case CriticalLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Fatal("%s", message)
			} else {
				logger.Fatal("%s %s", icon, message)
			}
		case DebugLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Debug("%s", message)
			} else {
				logger.Debug("%s %s", icon, message)
			}
		case SuccessLevel:
			if indentation != "" {
				icon := fmt.Sprintf("%s%s", indentation, icon)
				fmt.Print(icon)
				logger.Success("%s", message)
			} else {
				logger.Success("%s %s", icon, message)
			}
		}
	}
}

func (svc *NotificationsService) HasErrors() bool {
	for _, n := range svc.notifications.Items {
		if n.Level == CriticalLevel || n.Level == ErrorLevel && n.State == NewState {
			return true
		}
	}
	return false
}

func (svc *NotificationsService) CountErrors() uint {
	count := 0
	for _, n := range svc.notifications.Items {
		if n.Level == CriticalLevel || n.Level == ErrorLevel && n.State == NewState {
			count += 1
		}
	}
	return uint(count)
}

func (svc *NotificationsService) HasWarning() bool {
	for _, n := range svc.notifications.Items {
		if n.Level == WarningLevel && n.State == NewState {
			return true
		}
	}
	return false
}

func (svc *NotificationsService) CountWarnings() uint {
	count := 0
	for _, n := range svc.notifications.Items {
		if n.Level == WarningLevel && n.State == NewState {
			count += 1
		}
	}
	return uint(count)
}

func (svc *NotificationsService) HasCritical() bool {
	for _, n := range svc.notifications.Items {
		if n.Level == CriticalLevel && n.State == NewState {
			return true
		}
	}
	return false
}
