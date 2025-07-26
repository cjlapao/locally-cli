package certificates

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/services/executer"
	"github.com/cjlapao/locally-cli/internal/utils"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// certutil -enterprise -f -v -AddStore \"Root\"  " + config.baseDir + name + ".crt

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

func (i Installer) InstallCertificate(ctx *appctx.AppContext, filepath string, store CertificateStore) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate")
	os := utils.GetOperatingSystem()
	ctx.Log().Debugf("Starting to install certificate %v on %v", filepath, os.String())
	switch os {
	case utils.LinuxOperatingSystem:
		ctx.Log().Debug("Not implemented yet")
	case utils.WindowsOperatingSystem:
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

	return diag
}
