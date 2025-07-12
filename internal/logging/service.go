// Package logging provides a global logger for the application.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance
var Logger *logrus.Logger

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
	LogLevelPanic   LogLevel = "panic"
)

// Initialize sets up the global logger with configuration from the config service
func Initialize() {
	configService := config.GetInstance().Get()
	Logger = logrus.New()

	// Set default level
	level := getLogLevel(configService)
	isDebug := configService.IsDebug()
	Logger.SetLevel(level)

	// Enable caller reporting to show file and line numbers
	Logger.SetReportCaller(true)

	// Set custom formatter with caller information
	Logger.SetFormatter(&CustomFormatter{
		TextFormatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		},
		ShowCaller: true,
		IsDebug:    isDebug,
	})

	// Set output to stdout
	Logger.SetOutput(os.Stdout)

	// Log the initialization
	Logger.WithFields(logrus.Fields{
		"level": level.String(),
	}).Info("Logging service initialized")
}

// getLogLevel retrieves the log level from config service
func getLogLevel(cfg *config.Config) logrus.Level {
	if cfg == nil {
		return logrus.InfoLevel
	}

	levelStr := cfg.GetString(config.LogLevelKey, "info")
	if levelStr == "" {
		return logrus.InfoLevel
	}
	// If debug is enabled, set the log level to debug
	if cfg.IsDebug() {
		levelStr = "debug"
	}

	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warning", "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

// SetLevel changes the log level dynamically
func SetLevel(level LogLevel) {
	if Logger == nil {
		return
	}

	var logrusLevel logrus.Level
	switch level {
	case LogLevelDebug:
		logrusLevel = logrus.DebugLevel
	case LogLevelInfo:
		logrusLevel = logrus.InfoLevel
	case LogLevelWarning:
		logrusLevel = logrus.WarnLevel
	case LogLevelError:
		logrusLevel = logrus.ErrorLevel
	case LogLevelFatal:
		logrusLevel = logrus.FatalLevel
	case LogLevelPanic:
		logrusLevel = logrus.PanicLevel
	default:
		logrusLevel = logrus.InfoLevel
	}

	Logger.SetLevel(logrusLevel)
	Logger.WithField("level", level).Info("Log level changed")
}

func Debug(args ...interface{}) {
	if Logger != nil {
		Logger.Debug(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debugf(format, args...)
	}
}

func Info(args ...interface{}) {
	if Logger != nil {
		Logger.Info(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(format, args...)
	}
}

func Warn(args ...interface{}) {
	if Logger != nil {
		Logger.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(format, args...)
	}
}

func Error(args ...interface{}) {
	if Logger != nil {
		Logger.Error(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	if Logger != nil {
		Logger.Fatal(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Fatalf(format, args...)
	}
}

func Panic(args ...interface{}) {
	if Logger != nil {
		Logger.Panic(args...)
	}
}

func Panicf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Panicf(format, args...)
	}
}

func Trace(args ...interface{}) {
	if Logger != nil {
		Logger.Trace(args...)
	}
}

func Tracef(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Tracef(format, args...)
	}
}

// WithField creates a new entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	if Logger != nil {
		return Logger.WithField(key, value)
	}
	return nil
}

// WithFields creates a new entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	if Logger != nil {
		return Logger.WithFields(fields)
	}
	return nil
}

// WithError creates a new entry with an error field
func WithError(err error) *logrus.Entry {
	if Logger != nil {
		return Logger.WithError(err)
	}
	return nil
}

// Helper functions that automatically add file/line context

// DebugWithContext logs debug message with automatic file/line context
func DebugWithContext(msg string, args ...interface{}) {
	if Logger != nil {
		_, file, line, _ := runtime.Caller(1)
		filename := filepath.Base(file)
		Logger.WithField("location", fmt.Sprintf("%s:%d", filename, line)).Debugf(msg, args...)
	}
}

// InfoWithContext logs info message with automatic file/line context
func InfoWithContext(msg string, args ...interface{}) {
	if Logger != nil {
		_, file, line, _ := runtime.Caller(1)
		filename := filepath.Base(file)
		Logger.WithField("location", fmt.Sprintf("%s:%d", filename, line)).Infof(msg, args...)
	}
}

// WarnWithContext logs warning message with automatic file/line context
func WarnWithContext(msg string, args ...interface{}) {
	if Logger != nil {
		_, file, line, _ := runtime.Caller(1)
		filename := filepath.Base(file)
		Logger.WithField("location", fmt.Sprintf("%s:%d", filename, line)).Warnf(msg, args...)
	}
}

// ErrorWithContext logs error message with automatic file/line context
func ErrorWithContext(msg string, args ...interface{}) {
	if Logger != nil {
		_, file, line, _ := runtime.Caller(1)
		filename := filepath.Base(file)
		Logger.WithField("location", fmt.Sprintf("%s:%d", filename, line)).Errorf(msg, args...)
	}
}

// WithContext creates a new entry with automatic file/line context
func WithContext() *logrus.Entry {
	if Logger != nil {
		_, file, line, _ := runtime.Caller(1)
		filename := filepath.Base(file)
		return Logger.WithField("location", fmt.Sprintf("%s:%d", filename, line))
	}
	return nil
}

// CustomFormatter provides enhanced formatting with file/line information
type CustomFormatter struct {
	*logrus.TextFormatter
	ShowCaller bool
	IsDebug    bool
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Build the log line manually to avoid duplication
	var b strings.Builder

	// Add timestamp
	if f.FullTimestamp {
		b.WriteString(entry.Time.Format(f.TimestampFormat))
	} else {
		b.WriteString(entry.Time.Format("15:04:05"))
	}

	// Add level with color
	level := strings.ToUpper(entry.Level.String())
	b.WriteString(" level=")
	if f.ForceColors {
		b.WriteString(f.getLevelColor(entry.Level))
		b.WriteString(level)
		b.WriteString("\033[0m") // Reset color
	} else {
		b.WriteString(level)
	}

	// Add caller information if enabled
	if f.ShowCaller && entry.Caller != nil {
		// Get the full path and extract the part from the project root
		fullPath := entry.Caller.File
		identifier := extractProjectPath(fullPath)

		if f.ForceColors {
			b.WriteString("\033[35m") // Magenta for caller info
			if f.IsDebug {
				b.WriteString(fmt.Sprintf(" service=%s line=%d", identifier, entry.Caller.Line))
			} else {
				b.WriteString(fmt.Sprintf(" service=%s", identifier))
			}
			b.WriteString("\033[0m") // Reset color
		} else {
			if f.IsDebug {
				b.WriteString(fmt.Sprintf(" service=%s line=%d", identifier, entry.Caller.Line))
			} else {
				b.WriteString(fmt.Sprintf(" service=%s", identifier))
			}
		}
	}

	// Add message
	if entry.Message != "" {
		b.WriteString(" msg=\"")
		b.WriteString(entry.Message)
		b.WriteString("\"")
	}

	// Add fields with colors
	if len(entry.Data) > 0 {
		for key, value := range entry.Data {
			b.WriteString(" ")
			if f.ForceColors {
				b.WriteString("\033[36m") // Cyan for field keys
				b.WriteString(key)
				b.WriteString("\033[0m") // Reset color
				b.WriteString("=")
				b.WriteString("\033[37m") // White for field values
				b.WriteString(fmt.Sprintf("%v", value))
				b.WriteString("\033[0m") // Reset color
			} else {
				b.WriteString(key)
				b.WriteString("=")
				b.WriteString(fmt.Sprintf("%v", value))
			}
		}
	}

	b.WriteString("\n")

	return []byte(b.String()), nil
}

// getLevelColor returns the ANSI color code for the given log level
func (f *CustomFormatter) getLevelColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "\033[37m" // White
	case logrus.InfoLevel:
		return "\033[32m" // Green
	case logrus.WarnLevel:
		return "\033[33m" // Yellow
	case logrus.ErrorLevel:
		return "\033[31m" // Red
	case logrus.FatalLevel:
		return "\033[35m" // Magenta
	case logrus.PanicLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Default
	}
}

// extractClassName extracts the struct/class name from a Go function name
func extractClassName(functionName string) string {
	// Function name format: github.com/user/project/path.(*StructName).MethodName
	// or: github.com/user/project/path.StructName.MethodName
	// or: github.com/user/project/path.functionName

	// Split by dots
	parts := strings.Split(functionName, ".")
	if len(parts) < 2 {
		return filepath.Base(functionName) // Fallback to just the function name
	}

	// Get the last part (which contains the struct name and method)
	lastPart := parts[len(parts)-1]

	// Check if it's a method (contains parentheses for pointer receivers)
	if strings.Contains(lastPart, "(") {
		// Format: (*StructName).MethodName or (StructName).MethodName
		// Extract the struct name
		start := strings.Index(lastPart, "(") + 1
		end := strings.Index(lastPart, ")")
		if start > 0 && end > start {
			structPart := lastPart[start:end]
			// Remove the * if it's a pointer receiver
			structPart = strings.TrimPrefix(structPart, "*")
			return structPart
		}
	}

	// If it's not a method, it might be a package-level function
	// Try to get a meaningful name from the package path
	if len(parts) >= 2 {
		packageName := filepath.Base(parts[len(parts)-2])
		return packageName
	}

	return lastPart
}

// extractProjectPath extracts the path from the project root and removes the file extension
func extractProjectPath(fullPath string) string {
	// Look for common project root indicators
	projectRoots := []string{"lxc-agent", "internal", "cmd", "pkg"}

	for _, root := range projectRoots {
		if idx := strings.Index(fullPath, root); idx != -1 {
			// Extract from the project root onwards
			projectPath := fullPath[idx:]
			// Remove the file extension
			return strings.TrimSuffix(projectPath, filepath.Ext(projectPath))
		}
	}

	// Fallback: just return the filename without extension
	filename := filepath.Base(fullPath)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
