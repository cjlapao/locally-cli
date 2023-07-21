package certificates

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/cjlapao/common-go/helper"
)

//certutil -enterprise -f -v -AddStore \"Root\"  " + config.baseDir + name + ".crt

type Installer struct{}

type CertificateStore int

func (c CertificateStore) String() string {
	switch c {
	case RootStore:
		return "Root"
	case CAStore:
		return "CA"
	case WebHosting:
		return "WebHosting"
	default:
		return "WebHosting"
	}
}

const (
	RootStore CertificateStore = iota
	CAStore
	WebHosting
)

func (i Installer) InstallCertificate(filepath string, store CertificateStore) {
	os := helper.GetOperatingSystem()
	logger.Debug("Starting to install certificate %v on %v", filepath, os.String())
	switch os {
	case helper.LinuxOs:
		logger.Debug("Not implemented yet")
	case helper.WindowsOs:
		output, err := execute("certutil", "-enterprise", "-f", "-v", "-AddStore", store.String(), filepath)
		if err != nil {
			logger.Error("there was an error running root install on %v", filepath)
			logger.Error(output)
		}

		logger.Debug(output)
	}
}

func execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	var err bytes.Buffer
	cmd.Stderr = &err
	var in bytes.Buffer
	cmd.Stdin = &in

	cmd.Start()
	cmd.Wait()

	errString := err.String()
	if errString == "" && cmd.ProcessState.ExitCode() > 0 {
		errString = out.String()
	}

	if len(errString) > 0 {
		return out.String(), errors.New(errString)
	}

	return out.String(), nil
}
