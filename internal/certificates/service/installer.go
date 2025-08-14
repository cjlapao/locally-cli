package service

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/services/executer"
	"github.com/cjlapao/locally-cli/internal/utils"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// certutil -enterprise -f -v -AddStore \"Root\"  " + config.baseDir + name + ".crt

type CertificateInstaller struct{}

func NewCertificateInstaller() *CertificateInstaller {
	return &CertificateInstaller{}
}

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

func (i *CertificateInstaller) InstallCertificate(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate")
	os := utils.GetOperatingSystem()
	ctx.Log().Debugf("Starting to install certificate %v on %v", filepath, os.String())
	switch os {
	case utils.LinuxOperatingSystem:
		return i.installOnLinux(ctx, filepath, store)
	case utils.WindowsOperatingSystem:
		return i.installOnWindows(ctx, filepath, store)
	case utils.MacOSOperatingSystem:
		return i.installOnMacOS(ctx, filepath, store)
	}

	return diag
}

func (i *CertificateInstaller) UninstallCertificate(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("uninstall_certificate")
	os := utils.GetOperatingSystem()
	ctx.Log().Debugf("Starting to uninstall certificate %v on %v", filepath, os.String())
	switch os {
	case utils.LinuxOperatingSystem:
		return i.uninstallOnLinux(ctx, filepath, store)
	case utils.WindowsOperatingSystem:
		return i.uninstallOnWindows(ctx, filepath, store)
	case utils.MacOSOperatingSystem:
		return i.uninstallOnMacOS(ctx, filepath, store)
	}

	return diag
}

func (i *CertificateInstaller) installOnWindows(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate_on_windows")
	output, err := executer.ExecuteSimple(ctx, "certutil", "-enterprise", "-f", "-v", "-AddStore", store.String(), filepath)
	if err.HasErrors() {
		ctx.LogWithFields(map[string]interface{}{
			"filepath": filepath,
			"store":    store.String(),
			"output":   output.StdOut,
		}).Error("there was an error installing the certificate")
		diag.Append(err)
		return diag
	}

	diag.AddPathEntry("install_certificate", "installer", map[string]interface{}{
		"filepath": filepath,
		"store":    store.String(),
		"output":   output.StdOut,
	})

	ctx.LogWithFields(map[string]interface{}{
		"filepath": filepath,
		"store":    store.String(),
		"output":   output.StdOut,
	}).Info("certificate installed successfully")

	return diag
}

func (i *CertificateInstaller) uninstallOnWindows(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("uninstall_certificate_on_windows")
	output, err := executer.ExecuteSimple(ctx, "certutil", "-enterprise", "-f", "-v", "-RemoveStore", store.String(), filepath)
	if err.HasErrors() {
		ctx.LogWithFields(map[string]interface{}{
			"filepath": filepath,
			"store":    store.String(),
			"output":   output.StdOut,
		}).Error("there was an error uninstalling the certificate")
		diag.Append(err)
		return diag
	}

	diag.AddPathEntry("uninstall_certificate", "installer", map[string]interface{}{
		"filepath": filepath,
		"store":    store.String(),
		"output":   output.StdOut,
	})

	return diag
}

// Installs a certificate on Linux
func (i *CertificateInstaller) installOnLinux(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate_on_linux")
	ctx.Log().Debug("Not implemented yet")
	return diag
}

func (i *CertificateInstaller) uninstallOnLinux(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("uninstall_certificate_on_linux")
	ctx.Log().Debug("Not implemented yet")
	return diag
}

func (i *CertificateInstaller) installOnMacOS(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate_on_macos")
	ctx.Log().Debug("Not implemented yet")
	return diag
}

func (i CertificateInstaller) uninstallOnMacOS(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("uninstall_certificate_on_macos")
	ctx.Log().Debug("Not implemented yet")
	return diag
}
