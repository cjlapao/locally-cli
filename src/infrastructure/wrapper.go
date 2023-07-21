package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/icons"
	"github.com/cjlapao/locally-cli/notifications"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

type TerraformCommandWrapper struct {
	ToolPath      string
	CommandOutput string
	notifications *notifications.NotificationsService
}

func GetWrapper() *TerraformCommandWrapper {
	config = configuration.Get()
	return &TerraformCommandWrapper{
		notifications: notifications.Get(),
	}
}

func (svc *TerraformCommandWrapper) Init(path string, args ...string) error {
	notify.Rocket("Running Terraform Init on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "init")
	if helper.GetFlagSwitch("reconfigure", false) {
		runArgs = append(runArgs, "-reconfigure")
	}
	if helper.GetFlagSwitch("clean", false) {
		svc.cleanupFolder(path)
	}

	runArgs = append(runArgs, args...)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return changeDirErr
	}

	if config.Debug() {
		notify.Debug("Init Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return changeDirErr
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform init")
		notify.Error(output.StdErr)
		return err
	}

	if strings.Contains(output.GetAllOutput(), "Terraform initialized in an empty directory!") {
		err := errors.New("folder does not contain any terraform module to initiate")
		return err
	}

	if strings.Contains(output.GetAllOutput(), "Error:") {
		err := errors.New("something wrong while initiating the folder")
		return err
	}

	svc.CommandOutput = output.GetAllOutput()

	return nil
}

func (svc *TerraformCommandWrapper) Validate(path string, args ...string) error {
	notify.Rocket("Running Terraform validate on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "validate")
	runArgs = append(runArgs, "--json")
	if helper.GetFlagSwitch("reconfigure", false) {
		runArgs = append(runArgs, "-reconfigure")
	}
	if helper.GetFlagSwitch("upgrade", false) {
		runArgs = append(runArgs, "-upgrade=true")
	}
	runArgs = append(runArgs, args...)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return changeDirErr
	}

	if config.Debug() {
		notify.Debug("Validate Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, execErr := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	var result TerraformValidateResult
	if err := json.Unmarshal([]byte(output.StdOut), &result); err != nil {
		return err
	}

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return changeDirErr
	}

	if execErr != nil {
		if config.Debug() {
			notify.Debug("Validate output: %s", output)
		}
		for _, diagnostic := range result.Diagnostics {
			switch diagnostic.Severity {
			case "error":
				notify.Error("%s", diagnostic.Summary)
				notify.Error("    %s", diagnostic.Detail)
				notify.Error("    File: %s on line %s, column: %s", diagnostic.Range.Filename, fmt.Sprintf("%d", diagnostic.Range.Start.Line), fmt.Sprintf("%d", diagnostic.Range.Start.Column))
			}
		}
		notify.FromError(execErr, "Something wrong running terraform validate")
		return execErr
	}

	if !result.Valid {
		err := fmt.Errorf("there was an error validating path %v", path)
		notify.Info(output.GetAllOutput())
		return err
	}

	svc.CommandOutput = output.StdOut
	return nil
}

func (svc *TerraformCommandWrapper) Plan(path, outFilePath string, args ...string) (*PlanChanges, error) {
	notify.Rocket("Running Terraform plan on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "plan")
	runArgs = append(runArgs, "-input=false")
	runArgs = append(runArgs, "-compact-warnings")
	if outFilePath != "" {
		runArgs = append(runArgs, fmt.Sprintf("-out=%s", outFilePath))
	}

	runArgs = append(runArgs, args...)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return nil, changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if config.Debug() {
		notify.Debug("Plan Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteAndWatch(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if strings.Contains(output.GetAllOutput(), "Initialization required") {
		initError := errors.New(MissingInitWhenApplying)
		return nil, initError
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform plan")
		return nil, err
	}

	result, err := svc.Show(path, outFilePath)
	svc.CommandOutput = output.GetAllOutput()

	return result, err
}

func (svc *TerraformCommandWrapper) Apply(path, planFilePath string, args ...string) error {
	notify.Rocket("Running Terraform apply on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	backupDir := filepath.Dir(planFilePath)
	fileName := strings.ReplaceAll(filepath.Base(planFilePath), ".plan", ".backup")
	backupFilePath := helper.JoinPath(backupDir, fileName)

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "apply")
	runArgs = append(runArgs, "-compact-warnings")
	runArgs = append(runArgs, fmt.Sprintf("-backup=%s", backupFilePath))

	if planFilePath != "" {
		runArgs = append(runArgs, planFilePath)
	}

	if helper.GetFlagSwitch("auto-approve", false) || helper.GetFlagSwitch("no-input", false) {
		runArgs = append(runArgs, "-auto-approve")
	}

	runArgs = append(runArgs, args...)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return changeDirErr
	}

	if config.Debug() {
		notify.Debug("Apply Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteAndWatch(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return changeDirErr
	}

	if strings.Contains(output.GetAllOutput(), "Initialization required") {
		initError := errors.New(MissingInitWhenApplying)
		return initError
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform apply")
		return err
	}

	svc.CommandOutput = output.GetAllOutput()

	return nil
}

func (svc *TerraformCommandWrapper) Destroy(path, backupFilePath string, args ...string) error {
	notify.Rocket("Running Terraform destroy on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "destroy")
	runArgs = append(runArgs, fmt.Sprintf("-backup=%s", backupFilePath))
	runArgs = append(runArgs, "-auto-approve")

	runArgs = append(runArgs, args...)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return changeDirErr
	}

	if config.Debug() {
		notify.Debug("Destroy Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteAndWatch(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return changeDirErr
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform destroy")
		return err
	}

	svc.CommandOutput = output.GetAllOutput()

	return nil
}

func (svc *TerraformCommandWrapper) Show(path, outFilePath string) (*PlanChanges, error) {
	var planChanges PlanChanges
	notify.Rocket("Running Terraform show on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "show")
	runArgs = append(runArgs, "-json")
	runArgs = append(runArgs, outFilePath)

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return nil, changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if config.Debug() {
		notify.Debug("Show Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform show")
		return nil, err
	}

	var terraformPlan TerraformPlan
	err = json.Unmarshal([]byte(output.GetAllOutput()), &terraformPlan)
	if err != nil {
		notify.FromError(err, "Could not read terraform plan")
		return nil, err
	}

	if config.Debug() {
		output := filepath.Dir(outFilePath)
		fileName := strings.ReplaceAll(filepath.Base(outFilePath), ".plan", ".plan.json")
		filePath := helper.JoinPath(output, fileName)
		jsonPlan, err := json.Marshal(terraformPlan)
		if err != nil {
			return nil, err
		}

		helper.WriteToFile(string(jsonPlan), filePath)
	}

	planChanges, err = svc.readPlan(terraformPlan)
	if err != nil {
		notify.FromError(err, "Error processing terraform plan")
		return nil, err
	}

	svc.CommandOutput = output.GetAllOutput()

	if planChanges.HasChanges() {
		message := ""
		if planChanges.CreateOps > 0 {
			if len(message) > 0 {
				message += ", "
			}
			message += fmt.Sprintf("Creating %s resources", strconv.Itoa(planChanges.CreateOps))
		}
		if planChanges.ChangeOps > 0 {
			if len(message) > 0 {
				message += ", "
			}
			message += fmt.Sprintf("Changing %s resources", strconv.Itoa(planChanges.ChangeOps))
		}
		if planChanges.DeleteOps > 0 {
			if len(message) > 0 {
				message += ", "
			}
			message += fmt.Sprintf("Destroying %s resources", strconv.Itoa(planChanges.DeleteOps))
		}
		if planChanges.NoOps > 0 {
			if len(message) > 0 {
				message += ", "
			}
			message += fmt.Sprintf("%s resources with no-ops", strconv.Itoa(planChanges.NoOps))
		}

		notify.Hammer("Found changes in plan: %s", message)
	} else {
		notify.InfoWithIcon(icons.IconThumbsUp, "No changes found in plan, your infrastructure is up to date.")
	}
	return &planChanges, nil
}

func (svc *TerraformCommandWrapper) Output(path string) (map[string]TerraformOutputVariable, error) {
	result := make(map[string]TerraformOutputVariable)
	notify.Hammer("Running Terraform output on %s...", path)
	os.Setenv("root_path", helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH))

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "output")
	runArgs = append(runArgs, "-json")

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return result, changeDirErr
	}

	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return result, changeDirErr
	}

	if config.Debug() {
		notify.Debug("Output Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return result, changeDirErr
	}

	if err != nil {
		notify.FromError(err, "Something wrong running terraform output")
		return result, err
	}

	if err := json.Unmarshal([]byte(output.GetAllOutput()), &result); err != nil {
		return result, err
	}

	if config.Debug() {
		notify.Debug("Output Values: %v", fmt.Sprintf("%v", result))
	}

	svc.CommandOutput = output.GetAllOutput()

	return result, nil
}

func (svc *TerraformCommandWrapper) Format(path string) error {
	if path == "" {
		return nil
	}

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "fmt")
	runArgs = append(runArgs, path)

	if config.Debug() {
		notify.Debug("Ftm Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	_, err := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	if config.Verbose() {
		notify.Info("Formatted successfully variable file %s", path)
	}

	if err != nil {
		notify.FromError(err, "Something wrong formatting the variable file %s", path)
		return err
	}

	return nil
}

func (svc *TerraformCommandWrapper) DependencyGraph(path string) (*TerraformGraph, error) {
	if path == "" {
		err := errors.New("path cannot be empty")
		return nil, err
	}

	runArgs := make([]string, 0)
	runArgs = append(runArgs, "graph")

	currentFolder, changeDirErr := os.Getwd()
	if changeDirErr != nil {
		return nil, changeDirErr
	}
	changeDirErr = os.Chdir(path)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if config.Debug() {
		notify.Debug("Graph Run Arguments: %v", fmt.Sprintf("%v", runArgs))
	}

	output, err := executer.ExecuteWithNoOutput(configuration.GetTerraformPath(), runArgs...)

	if config.Debug() {
		notify.Info(output.GetAllOutput())
	}

	changeDirErr = os.Chdir(currentFolder)
	if changeDirErr != nil {
		return nil, changeDirErr
	}

	if config.Verbose() {
		notify.Info("Graph generated successfully variable file %s", path)
	}

	if err != nil {
		notify.FromError(err, "Something wrong generating graph for stack file %s", path)
		return nil, err
	}

	graph := readGraph(output.GetAllOutput())
	return graph, nil
}

func (svc *TerraformCommandWrapper) readPlan(plan TerraformPlan) (PlanChanges, error) {
	result := PlanChanges{
		CreateOps: 0,
		ChangeOps: 0,
		NoOps:     0,
	}
	for _, resource := range plan.ResourceChanges {
		for _, action := range resource.Change.Actions {
			if action == Create {
				result.CreateOps += 1
			}
			if action == NoOp {
				result.NoOps += 1
			}
			if action == Delete {
				result.DeleteOps += 1
			}
			if action == Update {
				result.ChangeOps += 1
			}
		}
	}
	for _, module := range plan.OutputChanges {
		for _, action := range module.Actions {
			if action == Create {
				result.CreateOps += 1
			}

			if action == Delete {
				result.DeleteOps += 1
			}

			if action == Update {
				result.ChangeOps += 1
			}
		}
	}

	return result, nil
}

func (svc *TerraformCommandWrapper) cleanupFolder(path string) error {
	notify.Info("Cleaning up terraform folder %s", path)
	hclFile := helper.JoinPath(path, ".terraform.lock.hcl")
	tempFolder := helper.JoinPath(path, ".terraform")
	if helper.FileExists(hclFile) {
		notify.Info("Removing terraform lock file from %s", path)
		if err := helper.DeleteFile(hclFile); err != nil {
			return err
		}
	}

	if helper.DirectoryExists(tempFolder) {
		notify.Info("Removing terraform module folder from %s", path)
		if err := helper.DeleteAllFiles(tempFolder); err != nil {
			return err
		}
	}

	return nil
}
