package utils

import (
	"runtime"
	"strings"
)

// OperatingSystem enum
type OperatingSystem int

// Defines the operating system Enum
const (
	WindowsOperatingSystem OperatingSystem = iota
	LinuxOperatingSystem
	MacOSOperatingSystem
	UnknownOperatingSystem
)

func (o OperatingSystem) String() string {
	return [...]string{"Windows", "Linux", "MacOS", "Unknown"}[o]
}

type Architecture int

const (
	AMD64Architecture Architecture = iota
	ARM64Architecture
	UnknownArchitecture
)

func (a Architecture) String() string {
	return [...]string{"AMD64", "ARM64", "Unknown"}[a]
}

// GetOperatingSystem returns the operating system
/*
Get the operating system name and return it as an OperatingSystem constant.

Args:
	None
Returns:
	The operating system.
*/
func GetOperatingSystem() OperatingSystem {
	os := runtime.GOOS
	switch strings.ToLower(os) {
	case "linux":
		return LinuxOperatingSystem
	case "windows":
		return WindowsOperatingSystem
	case "darwin":
		return MacOSOperatingSystem
	}
	return UnknownOperatingSystem
}

func GetArchitecture() Architecture {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return AMD64Architecture
	case "arm64":
		return ARM64Architecture
	}
	return UnknownArchitecture
}
