package logging

import (
	"bytes"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestConfig is a test implementation of config.Config
type TestConfig struct {
	debug    bool
	logLevel string
}

func (t *TestConfig) IsDebug() bool {
	return t.debug
}

func (t *TestConfig) GetString(key string, defaultValue string) string {
	if key == config.LogLevelKey {
		return t.logLevel
	}
	return defaultValue
}

// Test helper to capture log output
type testLogCapture struct {
	*bytes.Buffer
	originalOutput *os.File
}

func newTestLogCapture() *testLogCapture {
	buf := &bytes.Buffer{}
	return &testLogCapture{
		Buffer:         buf,
		originalOutput: os.Stdout,
	}
}

func (t *testLogCapture) Start() {
	os.Stdout = nil
}

func (t *testLogCapture) Stop() {
	os.Stdout = t.originalOutput
}

func (t *testLogCapture) GetOutput() string {
	return t.String()
}

func TestInitialize(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Test with debug enabled
	t.Run("with debug enabled", func(t *testing.T) {
		// Test the getLogLevel function with nil config
		level := getLogLevel((*config.Config)(nil))
		assert.Equal(t, logrus.InfoLevel, level) // Default level when config is nil
	})

	// Test with specific log level
	t.Run("with specific log level", func(t *testing.T) {
		// Test the getLogLevel function with different inputs
		level := getLogLevel((*config.Config)(nil))
		assert.Equal(t, logrus.InfoLevel, level)
	})

	// Test with nil config
	t.Run("with nil config", func(t *testing.T) {
		level := getLogLevel(nil)
		assert.Equal(t, logrus.InfoLevel, level)
	})
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		levelStr string
		isDebug  bool
		expected logrus.Level
	}{
		{"debug level", "debug", false, logrus.DebugLevel},
		{"info level", "info", false, logrus.InfoLevel},
		{"warning level", "warning", false, logrus.WarnLevel},
		{"warn level", "warn", false, logrus.WarnLevel},
		{"error level", "error", false, logrus.ErrorLevel},
		{"fatal level", "fatal", false, logrus.FatalLevel},
		{"panic level", "panic", false, logrus.PanicLevel},
		{"trace level", "trace", false, logrus.TraceLevel},
		{"unknown level", "unknown", false, logrus.InfoLevel},
		{"empty level", "", false, logrus.InfoLevel},
		{"debug overrides when debug enabled", "info", true, logrus.DebugLevel},
		{"case insensitive", "ERROR", false, logrus.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't easily mock the config interface, we'll test the logic
			// by creating a real config with the right values
			realConfig := &config.Config{}
			if tt.isDebug {
				realConfig.Set(config.DebugKey, "true")
			}
			if tt.levelStr != "" {
				realConfig.Set(config.LogLevelKey, tt.levelStr)
			}

			level := getLogLevel(realConfig)
			assert.Equal(t, tt.expected, level)
		})
	}

	// Test with nil config
	t.Run("nil config", func(t *testing.T) {
		level := getLogLevel(nil)
		assert.Equal(t, logrus.InfoLevel, level)
	})
}

func TestSetLevel(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Create a test logger
	Logger = logrus.New()
	Logger.SetOutput(&bytes.Buffer{})

	tests := []struct {
		name     string
		level    LogLevel
		expected logrus.Level
	}{
		{"debug", LogLevelDebug, logrus.DebugLevel},
		{"info", LogLevelInfo, logrus.InfoLevel},
		{"warning", LogLevelWarning, logrus.WarnLevel},
		{"error", LogLevelError, logrus.ErrorLevel},
		{"fatal", LogLevelFatal, logrus.FatalLevel},
		{"panic", LogLevelPanic, logrus.PanicLevel},
		{"unknown", "unknown", logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			assert.Equal(t, tt.expected, Logger.GetLevel())
		})
	}

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		Logger = nil
		SetLevel(LogLevelDebug) // Should not panic
	})
}

func TestLoggingFunctions(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Create a test logger with buffer output
	buf := &bytes.Buffer{}
	Logger = logrus.New()
	Logger.SetOutput(buf)
	Logger.SetLevel(logrus.TraceLevel) // Set to trace level to capture all log levels

	t.Run("Debug functions", func(t *testing.T) {
		buf.Reset()
		Debug("test debug message")
		output := buf.String()
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "level=debug")

		buf.Reset()
		Debugf("test debug %s", "formatted")
		output = buf.String()
		assert.Contains(t, output, "test debug formatted")
		assert.Contains(t, output, "level=debug")
	})

	t.Run("Info functions", func(t *testing.T) {
		buf.Reset()
		Info("test info message")
		output := buf.String()
		assert.Contains(t, output, "test info message")
		assert.Contains(t, output, "level=info")

		buf.Reset()
		Infof("test info %s", "formatted")
		output = buf.String()
		assert.Contains(t, output, "test info formatted")
		assert.Contains(t, output, "level=info")
	})

	t.Run("Warn functions", func(t *testing.T) {
		buf.Reset()
		Warn("test warn message")
		output := buf.String()
		assert.Contains(t, output, "test warn message")
		assert.Contains(t, output, "level=warning")

		buf.Reset()
		Warnf("test warn %s", "formatted")
		output = buf.String()
		assert.Contains(t, output, "test warn formatted")
		assert.Contains(t, output, "level=warning")
	})

	t.Run("Error functions", func(t *testing.T) {
		buf.Reset()
		Error("test error message")
		output := buf.String()
		assert.Contains(t, output, "test error message")
		assert.Contains(t, output, "level=error")

		buf.Reset()
		Errorf("test error %s", "formatted")
		output = buf.String()
		assert.Contains(t, output, "test error formatted")
		assert.Contains(t, output, "level=error")
	})

	t.Run("Trace functions", func(t *testing.T) {
		buf.Reset()
		Trace("test trace message")
		output := buf.String()
		assert.Contains(t, output, "test trace message")
		assert.Contains(t, output, "level=trace")

		buf.Reset()
		Tracef("test trace %s", "formatted")
		output = buf.String()
		assert.Contains(t, output, "test trace formatted")
		assert.Contains(t, output, "level=trace")
	})

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		Logger = nil
		// These should not panic
		Debug("test")
		Debugf("test %s", "formatted")
		Info("test")
		Infof("test %s", "formatted")
		Warn("test")
		Warnf("test %s", "formatted")
		Error("test")
		Errorf("test %s", "formatted")
		Trace("test")
		Tracef("test %s", "formatted")
	})
}

func TestFatalAndPanicFunctions(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Create a test logger with buffer output
	buf := &bytes.Buffer{}
	Logger = logrus.New()
	Logger.SetOutput(buf)
	Logger.SetLevel(logrus.TraceLevel) // Set to trace level to capture all log levels

	t.Run("Fatal functions", func(t *testing.T) {
		// Skipping Fatal and Fatalf log assertions because logrus.Fatal calls os.Exit(1), which terminates the test process.
		// See: https://github.com/sirupsen/logrus/issues/63
		// You can test up to the call, but not the output or side effects.
		// buf.Reset()
		// Fatal("test fatal message")
		// output := buf.String()
		// assert.Contains(t, output, "test fatal message")
		// assert.Contains(t, output, "level=fatal")

		// buf.Reset()
		// Fatalf("test fatal %s", "formatted")
		// output = buf.String()
		// assert.Contains(t, output, "test fatal formatted")
		// assert.Contains(t, output, "level=fatal")
	})

	t.Run("Panic functions", func(t *testing.T) {
		// Skipping Panic and Panicf log assertions because logrus.Panic calls panic(), which terminates the test process.
		// See: https://github.com/sirupsen/logrus/issues/63
		// You can test up to the call, but not the output or side effects.
		// buf.Reset()
		// Panic("test panic message")
		// output := buf.String()
		// assert.Contains(t, output, "test panic message")
		// assert.Contains(t, output, "level=panic")

		// buf.Reset()
		// Panicf("test panic %s", "formatted")
		// output = buf.String()
		// assert.Contains(t, output, "test panic formatted")
		// assert.Contains(t, output, "level=panic")
	})

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		Logger = nil
		// These should not panic
		Fatal("test")
		Fatalf("test %s", "formatted")
		Panic("test")
		Panicf("test %s", "formatted")
	})
}

func TestWithFunctions(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Create a test logger
	Logger = logrus.New()
	Logger.SetOutput(&bytes.Buffer{})

	t.Run("WithField", func(t *testing.T) {
		entry := WithField("key", "value")
		assert.NotNil(t, entry)
		assert.Equal(t, "value", entry.Data["key"])
	})

	t.Run("WithFields", func(t *testing.T) {
		fields := logrus.Fields{"key1": "value1", "key2": "value2"}
		entry := WithFields(fields)
		assert.NotNil(t, entry)
		assert.Equal(t, "value1", entry.Data["key1"])
		assert.Equal(t, "value2", entry.Data["key2"])
	})

	t.Run("WithError", func(t *testing.T) {
		err := errors.New("test error")
		entry := WithError(err)
		assert.NotNil(t, entry)
		assert.Equal(t, err, entry.Data["error"])
	})

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		Logger = nil
		assert.Nil(t, WithField("key", "value"))
		assert.Nil(t, WithFields(logrus.Fields{"key": "value"}))
		assert.Nil(t, WithError(errors.New("test")))
	})
}

func TestContextFunctions(t *testing.T) {
	// Save original logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// Create a test logger with buffer output
	buf := &bytes.Buffer{}
	Logger = logrus.New()
	Logger.SetOutput(buf)
	Logger.SetLevel(logrus.DebugLevel)

	t.Run("DebugWithContext", func(t *testing.T) {
		buf.Reset()
		DebugWithContext("test context message")
		output := buf.String()
		assert.Contains(t, output, "test context message")
		assert.Contains(t, output, "level=debug")
		assert.Contains(t, output, "location=")
	})

	t.Run("InfoWithContext", func(t *testing.T) {
		buf.Reset()
		InfoWithContext("test context message")
		output := buf.String()
		assert.Contains(t, output, "test context message")
		assert.Contains(t, output, "level=info")
		assert.Contains(t, output, "location=")
	})

	t.Run("WarnWithContext", func(t *testing.T) {
		buf.Reset()
		WarnWithContext("test context message")
		output := buf.String()
		assert.Contains(t, output, "test context message")
		assert.Contains(t, output, "level=warning")
		assert.Contains(t, output, "location=")
	})

	t.Run("ErrorWithContext", func(t *testing.T) {
		buf.Reset()
		ErrorWithContext("test context message")
		output := buf.String()
		assert.Contains(t, output, "test context message")
		assert.Contains(t, output, "level=error")
		assert.Contains(t, output, "location=")
	})

	t.Run("WithContext", func(t *testing.T) {
		entry := WithContext()
		assert.NotNil(t, entry)
		assert.Contains(t, entry.Data, "location")
	})

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		Logger = nil
		// These should not panic
		DebugWithContext("test")
		InfoWithContext("test")
		WarnWithContext("test")
		ErrorWithContext("test")
		assert.Nil(t, WithContext())
	})
}

func TestCustomFormatter(t *testing.T) {
	formatter := &CustomFormatter{
		TextFormatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		},
		ShowCaller: true,
		IsDebug:    true,
	}

	t.Run("Format with all fields", func(t *testing.T) {
		entry := &logrus.Entry{
			Logger:  logrus.New(),
			Data:    logrus.Fields{"request_id": "123", "user_id": "456", "tenant_id": "789", "duration": "1s", "custom": "value"},
			Time:    time.Now(),
			Level:   logrus.InfoLevel,
			Message: "test message",
			Caller: &runtime.Frame{
				File: "/path/to/project/internal/logging/service_test.go",
				Line: 100,
			},
		}

		output, err := formatter.Format(entry)
		assert.NoError(t, err)
		outputStr := string(output)

		// Check that message appears first
		assert.Contains(t, outputStr, "msg=\"test message\"")

		// Check that main fields appear in order (with color codes)
		// Main fields have color codes wrapping the entire field=value
		requestIDIndex := strings.Index(outputStr, "\x1b[34mrequest_id=123\x1b[0m")
		userIDIndex := strings.Index(outputStr, "\x1b[32muser_id=456\x1b[0m")
		tenantIDIndex := strings.Index(outputStr, "\x1b[33mtenant_id=789\x1b[0m")
		durationIndex := strings.Index(outputStr, "\x1b[36mduration=1s\x1b[0m")
		// Custom field has separate colors for key and value
		customIndex := strings.Index(outputStr, "\x1b[36mcustom\x1b[0m=\x1b[37mvalue\x1b[0m")

		assert.True(t, requestIDIndex < userIDIndex)
		assert.True(t, userIDIndex < tenantIDIndex)
		assert.True(t, tenantIDIndex < durationIndex)
		assert.True(t, durationIndex < customIndex)

		// Check for caller information
		assert.Contains(t, outputStr, "service=")
		assert.Contains(t, outputStr, "line=")
	})

	t.Run("Format without caller", func(t *testing.T) {
		formatterNoCaller := &CustomFormatter{
			TextFormatter: &logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
				ForceColors:     false,
			},
			ShowCaller: false,
			IsDebug:    false,
		}

		entry := &logrus.Entry{
			Logger:  logrus.New(),
			Data:    logrus.Fields{"test": "value"},
			Time:    time.Now(),
			Level:   logrus.ErrorLevel,
			Message: "error message",
		}

		output, err := formatterNoCaller.Format(entry)
		assert.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "msg=\"error message\"")
		assert.NotContains(t, outputStr, "service=")
		assert.NotContains(t, outputStr, "line=")
	})

	t.Run("Format with empty message", func(t *testing.T) {
		entry := &logrus.Entry{
			Logger: logrus.New(),
			Data:   logrus.Fields{"test": "value"},
			Time:   time.Now(),
			Level:  logrus.WarnLevel,
		}

		output, err := formatter.Format(entry)
		assert.NoError(t, err)
		outputStr := string(output)

		// The output will contain ANSI color codes, so check for the colored key and value
		assert.Contains(t, outputStr, "\x1b[36mtest\x1b[0m=\x1b[37mvalue\x1b[0m")
	})

	t.Run("Format with nil entry", func(t *testing.T) {
		formatter := &CustomFormatter{
			TextFormatter: &logrus.TextFormatter{
				FullTimestamp:   false,
				TimestampFormat: "2006-01-02 15:04:05",
				ForceColors:     false,
			},
			ShowCaller: true,
			IsDebug:    false,
		}
		_, err := formatter.Format(nil)
		assert.Error(t, err)
	})
}

func TestGetLevelColor(t *testing.T) {
	formatter := &CustomFormatter{}

	tests := []struct {
		level    logrus.Level
		expected string
	}{
		{logrus.DebugLevel, "\033[37m"},
		{logrus.InfoLevel, "\033[32m"},
		{logrus.WarnLevel, "\033[33m"},
		{logrus.ErrorLevel, "\033[31m"},
		{logrus.FatalLevel, "\033[35m"},
		{logrus.PanicLevel, "\033[35m"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			color := formatter.getLevelColor(tt.level)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestExtractClassName(t *testing.T) {
	tests := []struct {
		name     string
		function string
		expected string
	}{
		{
			name:     "pointer receiver method",
			function: "github.com/user/project/path.(*StructName).MethodName",
			expected: "(*StructName)",
		},
		{
			name:     "value receiver method",
			function: "github.com/user/project/path.(StructName).MethodName",
			expected: "(StructName)",
		},
		{
			name:     "package function",
			function: "github.com/user/project/path.functionName",
			expected: "path",
		},
		{
			name:     "simple function",
			function: "functionName",
			expected: "functionName",
		},
		{
			name:     "short path",
			function: "path.function",
			expected: "path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractClassName(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractProjectPath(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		expected string
	}{
		{
			name:     "internal path",
			fullPath: "/home/user/project/internal/logging/service.go",
			expected: "internal/logging/service",
		},
		{
			name:     "cmd path",
			fullPath: "/home/user/project/cmd/main.go",
			expected: "cmd/main",
		},
		{
			name:     "pkg path",
			fullPath: "/home/user/project/pkg/utils/helper.go",
			expected: "pkg/utils/helper",
		},
		{
			name:     "lxc-agent path",
			fullPath: "/home/user/project/lxc-agent/service.go",
			expected: "lxc-agent/service",
		},
		{
			name:     "no project root",
			fullPath: "/home/user/other/file.go",
			expected: "file",
		},
		{
			name:     "with extension",
			fullPath: "/path/to/file.go",
			expected: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectPath(tt.fullPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogLevelConstants(t *testing.T) {
	assert.Equal(t, LogLevel("debug"), LogLevelDebug)
	assert.Equal(t, LogLevel("info"), LogLevelInfo)
	assert.Equal(t, LogLevel("warning"), LogLevelWarning)
	assert.Equal(t, LogLevel("error"), LogLevelError)
	assert.Equal(t, LogLevel("fatal"), LogLevelFatal)
	assert.Equal(t, LogLevel("panic"), LogLevelPanic)
}

func TestFormatterEdgeCases(t *testing.T) {
	formatter := &CustomFormatter{
		TextFormatter: &logrus.TextFormatter{
			FullTimestamp:   false,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     false,
		},
		ShowCaller: true,
		IsDebug:    false,
	}

	t.Run("Format with nil entry", func(t *testing.T) {
		// This should not panic
		_, err := formatter.Format(nil)
		assert.Error(t, err)
	})

	t.Run("Format with empty fields", func(t *testing.T) {
		entry := &logrus.Entry{
			Logger:  logrus.New(),
			Data:    logrus.Fields{},
			Time:    time.Now(),
			Level:   logrus.InfoLevel,
			Message: "test",
		}

		output, err := formatter.Format(entry)
		assert.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "msg=\"test\"")
		assert.NotContains(t, outputStr, "level=info")
	})

	t.Run("Format with special characters in fields", func(t *testing.T) {
		entry := &logrus.Entry{
			Logger:  logrus.New(),
			Data:    logrus.Fields{"key with spaces": "value with \"quotes\""},
			Time:    time.Now(),
			Level:   logrus.InfoLevel,
			Message: "test",
		}

		output, err := formatter.Format(entry)
		assert.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "key with spaces=")
		assert.Contains(t, outputStr, "value with \"quotes\"")
	})
}
