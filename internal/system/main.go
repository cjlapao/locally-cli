package system

import (
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cjlapao/locally-cli/internal/caddy"
	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/icons"
	"github.com/cjlapao/locally-cli/internal/notifications"

	"github.com/cjlapao/common-go/execution_context"
	"github.com/cjlapao/common-go/helper"
)

type SystemService struct {
	BuildErrors   []error
	BuildWarnings []string
}

var (
	globalSystemService *SystemService
	notify              = notifications.Get()
	config              *configuration.ConfigService
	context             = execution_context.Get()
)

func Get() *SystemService {
	if globalSystemService != nil {
		return globalSystemService
	}

	return New()
}

func New() *SystemService {
	config = configuration.Get()
	globalSystemService = &SystemService{}

	return globalSystemService
}

func (svc *SystemService) Init() error {
	svc.setDefaultEnvVariables()
	svc.setupGracefulShutdown()
	svc.CheckFolders(false)

	return nil
}

func (svc *SystemService) CheckFolders(clean bool) {
	notify.Hammer("Checking for %s folders", common.SERVICE_NAME)
	folderPath := config.GetCurrentContext().Configuration.OutputPath
	if !helper.DirectoryExists(folderPath) {
		if helper.CreateDirectory(folderPath, fs.ModePerm) {
			notify.InfoWithIcon(icons.IconCheckMark, "Output folder created")
		} else {
			notify.Critical("There was an error creating the output folder")
		}
	}

	// This is to create the config-data for config service, this allows a more automated way
	configDataFolder := helper.JoinPath(folderPath, common.DEFAULT_CONFIG_SERVICE_PATH)
	if !helper.DirectoryExists(configDataFolder) {
		if helper.CreateDirectory(configDataFolder, fs.ModePerm) {
			notify.InfoWithIcon(icons.IconCheckMark, "Config service data folder created")
		} else {
			notify.Critical("There was an error creating the Config service data folder")
		}
	}

	caddyFolder := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH)
	webClientsFolder := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.SPA_PATH)
	infrastructure := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.INFRASTRUCTURE_PATH)
	pipelines := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.PIPELINES_PATH)

	if clean {
		// Cleaning the Caddy folder
		if helper.DirectoryExists(caddyFolder) {
			if err := helper.DeleteAllFiles(caddyFolder); err != nil {
				notify.Warning("There was an error cleaning the caddy folder")
			} else {
				notify.InfoWithIcon(icons.IconCheckMark, "Caddy folder was cleaned successfully")
			}
		}

		// Cleaning the webclients folder
		if helper.DirectoryExists(webClientsFolder) {
			if err := helper.DeleteAllFiles(webClientsFolder); err != nil {
				notify.Warning("There was an error cleaning the web clients folder")
			} else {
				notify.InfoWithIcon(icons.IconCheckMark, "Web clients folder was cleaned successfully")
			}
		}

		// Cleaning the Infrastructure folder
		if helper.DirectoryExists(infrastructure) {
			if err := helper.DeleteAllFiles(infrastructure); err != nil {
				notify.Warning("There was an error cleaning the infrastructure folder")
			} else {
				notify.InfoWithIcon(icons.IconCheckMark, "Infrastructure folder was cleaned successfully")
			}
		}

		// Cleaning the Pipelines folder
		if helper.DirectoryExists(pipelines) {
			if err := helper.DeleteAllFiles(pipelines); err != nil {
				notify.Warning("There was an error cleaning the pipelines folder")
			} else {
				notify.InfoWithIcon(icons.IconCheckMark, "Pipelines folder was cleaned successfully")
			}
		}
	}
}

func (svc *SystemService) CleanFolder(folder string) {
	folder = helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, folder)
	if helper.DirectoryExists(folder) {
		if err := helper.DeleteAllFiles(folder); err != nil {
			notify.Critical("There was an error cleaning the folder %s", folder)
		} else {
			notify.InfoWithIcon(icons.IconCheckMark, "Folder %s was cleaned successfully", folder)
		}
	}
}

func (svc *SystemService) setDefaultEnvVariables() {
	ctx := config.GetCurrentContext()
	if ctx.Configuration == nil {
		notify.Error("Configuration is not present, exiting")
		os.Exit(1)
	}

	// Root Path
	folderPath := config.GetCurrentContext().Configuration.OutputPath
	if folderPath == "" {
		contextFolderPath := context.Configuration.GetString("outputPath")
		if contextFolderPath != "" {
			config.GetCurrentContext().Configuration.OutputPath = helper.ToOsPath(contextFolderPath)
		} else {
			rootFilepath, err := os.Executable()
			if err != nil {
				notify.FromError(err, "Error getting the current filepath")
				config.GetCurrentContext().Configuration.OutputPath = helper.ToOsPath("./")
			}
			rootDir := filepath.Dir(rootFilepath)
			config.GetCurrentContext().Configuration.OutputPath = helper.ToOsPath(rootDir)
		}
	}

	notify.Hammer("Setting root path to %v", config.GetCurrentContext().Configuration.OutputPath)
}

func (svc *SystemService) setupGracefulShutdown() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		if sig == os.Interrupt || sig == os.Kill {
			caddySvc := caddy.GetWrapper()
			go caddySvc.Stop()
			os.Exit(2)
		}
	}()
}
