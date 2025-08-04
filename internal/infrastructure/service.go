package infrastructure

// import (
// 	"bufio"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io/fs"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/cjlapao/locally-cli/internal/azure_cli"
// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/context"
// 	context_entities "github.com/cjlapao/locally-cli/internal/context/entities"
// 	"github.com/cjlapao/locally-cli/internal/context/infrastructure_component"
// 	"github.com/cjlapao/locally-cli/internal/dependency_tree"
// 	"github.com/cjlapao/locally-cli/internal/entities"
// 	"github.com/cjlapao/locally-cli/internal/environment"
// 	"github.com/cjlapao/locally-cli/internal/executer"
// 	"github.com/cjlapao/locally-cli/internal/git"
// 	"github.com/cjlapao/locally-cli/internal/helpers"
// 	"github.com/cjlapao/locally-cli/internal/icons"
// 	"github.com/cjlapao/locally-cli/internal/mappers"
// 	"github.com/cjlapao/locally-cli/internal/notifications"

// 	cryptorand "github.com/cjlapao/common-go-cryptorand"
// 	"github.com/cjlapao/common-go/helper"
// 	"gopkg.in/yaml.v3"
// )

// var globalTerraformService *TerraformService

// type TerraformService struct {
// 	notify  *notifications.NotificationsService
// 	wrapper *TerraformCommandWrapper
// }

// type TerraformServiceOptions struct {
// 	Name              string
// 	BuildDependencies bool
// 	RootFolder        string
// 	StackPath         string
// 	StdOutput         bool
// }

// func New() *TerraformService {
// 	svc := TerraformService{
// 		wrapper: GetWrapper(),
// 		notify:  notifications.New(TerraformServiceName),
// 	}

// 	return &svc
// }

// func Get() *TerraformService {
// 	if globalTerraformService != nil {
// 		return globalTerraformService
// 	}

// 	return New()
// }

// func (svc *TerraformService) base(operation, name string) (*configuration.ConfigService, *context.Context) {
// 	if name != "" {
// 		notify.Wrench("%s infrastructure stack %s", operation, name)
// 	} else {
// 		notify.Wrench("%s infrastructure stacks", operation)
// 	}

// 	config := configuration.Get()
// 	context := config.GetCurrentContext()
// 	if context.Infrastructure == nil {
// 		notify.Warning("No infrastructure stacks found to initiate")
// 		return nil, nil
// 	}

// 	svc.setTerraformAuthContext(context)

// 	return config, context
// }

// func (svc *TerraformService) generateTerraformBackendConfig(context *context.Context, stack *infrastructure_component.InfrastructureStack) (string, error) {
// 	env := environment.GetInstance()
// 	notify.Wrench("Setting up %s stack backend configuration", stack.Name)
// 	outputPath := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)

// 	if !helper.FileExists(outputPath) {
// 		notify.Hammer("Creating %s folder", outputPath)
// 		if !helper.CreateDirectory(outputPath, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", outputPath)
// 			return "", err
// 		}
// 	}

// 	filePath := helper.JoinPath(outputPath, fmt.Sprintf("%s_backend_config.tf", common.EncodeName(stack.Name)))

// 	if !strings.HasSuffix(stack.Backend.StateFileName, ".tfstate") {
// 		stack.Backend.StateFileName = fmt.Sprintf("%s.tfstate", env.Replace(stack.Backend.StateFileName))
// 	}

// 	ctx := config.GetCurrentContext()
// 	if ctx.BackendConfig != nil && ctx.BackendConfig.Azure != nil {
// 		if stack.Backend == nil {
// 			stack.Backend = &infrastructure_component.InfrastructureAzureBackend{}
// 		}

// 		if stack.Backend.AccessKey == "" {
// 			stack.Backend.AccessKey = ctx.BackendConfig.Azure.AccessKey
// 		}
// 		if stack.Backend.ContainerName == "" {
// 			stack.Backend.ContainerName = ctx.BackendConfig.Azure.ContainerName
// 		}
// 		if stack.Backend.ResourceGroupName == "" {
// 			stack.Backend.ResourceGroupName = ctx.BackendConfig.Azure.ResourceGroupName
// 		}
// 		if stack.Backend.StorageAccountName == "" {
// 			stack.Backend.StorageAccountName = ctx.BackendConfig.Azure.StorageAccountName
// 		}
// 	}

// 	configContent := ""
// 	configContent += fmt.Sprintf("resource_group_name=\"%s\"\n", env.Replace(stack.Backend.ResourceGroupName))
// 	configContent += fmt.Sprintf("storage_account_name=\"%s\"\n", env.Replace(stack.Backend.StorageAccountName))
// 	configContent += fmt.Sprintf("container_name=\"%s\"\n", env.Replace(stack.Backend.ContainerName))
// 	configContent += fmt.Sprintf("key=\"%s\"\n", env.Replace(stack.Backend.StateFileName))
// 	configContent += fmt.Sprintf("access_key=\"%s\"\n", env.Replace(stack.Backend.AccessKey))

// 	if err := helper.WriteToFile(configContent, filePath); err != nil {
// 		return "", err
// 	}

// 	if common.IsDebug() {
// 		notify.Debug(configContent)
// 	}

// 	return filePath, nil
// }

// func (svc *TerraformService) generateTerraformVariableFile(context *context.Context, stack *infrastructure_component.InfrastructureStack) (string, error) {
// 	if stack == nil || len(stack.Variables) == 0 {
// 		err := fmt.Errorf("%s no variables found in %s stack", icons.IconInfo, stack.Name)
// 		return "", err
// 	}

// 	notify.Wrench("Setting up %s stack variable file", stack.Name)
// 	outputPath := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)

// 	if !helper.FileExists(outputPath) {
// 		notify.Hammer("Creating %s folder", outputPath)
// 		if !helper.CreateDirectory(outputPath, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", outputPath)
// 			return "", err
// 		}
// 	}

// 	filePath := helper.JoinPath(outputPath, fmt.Sprintf("%s_variables.tf", common.EncodeName(stack.Name)))
// 	if common.IsDebug() {
// 		notify.Debug("Variable File Path: %s", filePath)
// 	}

// 	variableContent := MarshalVariables(stack.Variables, "  ")
// 	if err := helper.WriteToFile(variableContent, filePath); err != nil {
// 		return "", err
// 	}

// 	svc.wrapper.Format(filePath)

// 	if common.IsDebug() {
// 		notify.Debug("Variable File Content:\n%v", variableContent)
// 	}

// 	return filePath, nil
// }

// func (svc *TerraformService) setTerraformAuthContext(context *context.Context) {
// 	currentTenant := os.Getenv("ARM_TENANT_ID")
// 	if context.Infrastructure.Authorization != nil {
// 		if !strings.EqualFold(currentTenant, context.Infrastructure.Authorization.TenantId) {
// 			logger.Info("%s Setting up authorization environment variables", icons.IconWrench)
// 			os.Setenv("ARM_CLIENT_ID", context.Infrastructure.Authorization.ClientId)
// 			os.Setenv("ARM_CLIENT_SECRET", context.Infrastructure.Authorization.ClientSecret)
// 			os.Setenv("ARM_SUBSCRIPTION_ID", context.Infrastructure.Authorization.SubscriptionId)
// 			os.Setenv("ARM_TENANT_ID", context.Infrastructure.Authorization.TenantId)
// 		}
// 	}
// }

// func (svc *TerraformService) CheckForTerraform(softFail bool) {
// 	config = configuration.Get()
// 	if !config.GlobalConfiguration.Tools.Checked.TerraformChecked {
// 		notify.Wrench("Checking for terraform tool in the system")
// 		if output, err := executer.ExecuteWithNoOutput(helpers.GetTerraformPath(), "version", "--json"); err != nil {
// 			if !softFail {
// 				notify.Error("Terraform tool not found in system, this is required for the selected function")
// 				os.Exit(1)
// 			} else {
// 				notify.Warning("Terraform tool not found in system, this might generate an error in the future")
// 			}
// 		} else {
// 			var jOutput TerraformVersion
// 			if err := json.Unmarshal([]byte(output.StdOut), &jOutput); err != nil {
// 				if !softFail {
// 					notify.Error("Terraform tool not found in system, this is required for the selected function")
// 					os.Exit(1)
// 				} else {
// 					notify.Warning("Terraform tool not found in system, this might generate an error in the future")
// 				}
// 			}

// 			if jOutput.TerraformRevision != "" {
// 				notify.Success("Terraform tool found with version %s and revision %s", jOutput.TerraformVersion, jOutput.TerraformRevision)
// 			} else {
// 				notify.Success("Terraform tool found with version %s", jOutput.TerraformVersion)
// 			}
// 		}
// 		config.GlobalConfiguration.Tools.Checked.TerraformChecked = true
// 	}
// }

// func (svc *TerraformService) CheckForCredentials() error {
// 	// TODO: adding agnostic way for the infrastructure
// 	ctx := config.GetCurrentContext()
// 	needsCreating := false
// 	if ctx.Credentials == nil {
// 		ctx.Credentials = &context_entities.Credentials{
// 			Azure: &entities.AzureCredentials{},
// 		}
// 		needsCreating = true
// 	}

// 	if ctx.Credentials.Azure.ClientId == "" || ctx.Credentials.Azure.ClientSecret == "" {
// 		needsCreating = true
// 	}

// 	if ctx.Credentials.Azure.SubscriptionId == "" || ctx.Credentials.Azure.TenantId == "" {
// 		notify.Warning("Could not find the necessary credentials to run infrastructure, please fill in in your context the credential object")
// 		err := fmt.Errorf("could not find a subscriptionId or tenantId in the configuration")
// 		return err
// 	}

// 	if needsCreating {
// 		notify.Hammer("Creating the required credentials to run infrastructure")
// 		name := fmt.Sprintf("locally-%s-%s", common.EncodeName(ctx.Name), cryptorand.GetRandomString(5))
// 		azCli := azure_cli.Get()
// 		notify.Wrench("Logging in with user to Azure to avoid issues")
// 		if ctx.Credentials == nil || ctx.Credentials.Azure == nil {
// 			return errors.New("credentials cannot be nil")
// 		}

// 		if err := azCli.UserLogin(ctx.Credentials.Azure.SubscriptionId, ctx.Credentials.Azure.TenantId); err != nil {
// 			return err
// 		}

// 		if _, err := azCli.CreateServicePrincipal(name, ctx.Credentials.Azure.SubscriptionId); err != nil {
// 			return err
// 		}
// 	}

// 	// Creating the authorization object if it does not exists
// 	if ctx.Infrastructure.Authorization == nil {
// 		ctx.Infrastructure.Authorization = &infrastructure_component.InfrastructureAuthorization{}
// 	}

// 	// Adding the variables into the right zone
// 	ctx.Infrastructure.Authorization.ClientId = ctx.Credentials.Azure.ClientId
// 	ctx.Infrastructure.Authorization.ClientSecret = ctx.Credentials.Azure.ClientSecret
// 	ctx.Infrastructure.Authorization.SubscriptionId = ctx.Credentials.Azure.SubscriptionId
// 	ctx.Infrastructure.Authorization.TenantId = ctx.Credentials.Azure.TenantId

// 	return nil
// }

// func (svc *TerraformService) InitBackendResources() error {
// 	ctx := config.GetCurrentContext()
// 	needsInitializing := false

// 	if ctx.Infrastructure == nil {
// 		err := errors.New("no infrastructure found, exiting")
// 		return err
// 	}

// 	if ctx.BackendConfig == nil {
// 		err := errors.New("no infrastructure backend configuration found, exiting")
// 		return err
// 	}

// 	if ctx.BackendConfig.Azure == nil {
// 		err := errors.New("no azure infrastructure backend configuration found, exiting")
// 		return err
// 	}

// 	notify.Debug("Init Backend Config: %s", fmt.Sprintf("%v", *ctx.BackendConfig))
// 	initBackend := helper.GetFlagSwitch("init-backend", false)
// 	if ctx.BackendConfig.LastInitiated == nil || initBackend {
// 		needsInitializing = true
// 	}

// 	if needsInitializing && ctx.BackendConfig.Azure != nil {
// 		azureCliSvc := azure_cli.Get()
// 		azureCliSvc.CheckForAzureCli(false)

// 		if err := azureCliSvc.InitBackendResources(ctx.BackendConfig.Azure); err != nil {
// 			return err
// 		}

// 		lastInitialized := time.Now()
// 		ctx.BackendConfig.LastInitiated = &lastInitialized
// 		if err := ctx.SaveBackendConfig(); err != nil {
// 			return err
// 		}

// 		return nil
// 	}

// 	notify.Success("Backend already initialized, continuing...")
// 	return nil
// }

// func (svc *TerraformService) InitiateStack(options *TerraformServiceOptions) {
// 	config, context := svc.base("Initiate", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	commandArgs := make([]string, 0)
// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)

// 	if options.BuildDependencies {
// 		var dependencyError error
// 		stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			return
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		if stack.Backend != nil {
// 			path, err := svc.generateTerraformBackendConfig(context, stack)
// 			if err != nil {
// 				notify.Error(err.Error())
// 				return
// 			}

// 			commandArgs = append(commandArgs, fmt.Sprintf("-backend-config=%s", path))
// 		}

// 		if stackPath == "" || !helper.DirectoryExists(stackPath) {
// 			err := errors.New("path does not exists")
// 			notify.FromError(err, "There was an error initiating the %s stack", stack.Name)
// 			return
// 		}

// 		if err := svc.wrapper.Init(stackPath, commandArgs...); err != nil {
// 			notify.Error(err.Error())
// 			return
// 		}

// 		notify.Success("Initiated successfully %s stack", stack.Name)
// 	}

// 	notify.Wrench("Attempting to Graph the stacks")
// 	svc.GraphStack(options)
// }

// func (svc *TerraformService) ValidateStack(options *TerraformServiceOptions) {
// 	config, context := svc.base("Validate", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	commandArgs := make([]string, 0)
// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)
// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	if options.BuildDependencies {
// 		var dependencyError error
// 		stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			return
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		if stackPath == "" || !helper.DirectoryExists(stackPath) {
// 			notify.Error("There was an error initiating the %s stack, err: path does not exists", stack.Name)
// 		}

// 		if err := svc.wrapper.Validate(stackPath, commandArgs...); err != nil {
// 			notify.Error(err.Error())
// 			return
// 		}

// 		notify.Success("Validated successfully %s stack", stack.Name)
// 	}
// }

// func (svc *TerraformService) PlanStack(options *TerraformServiceOptions) *PlanChanges {
// 	result := PlanChanges{}
// 	config, context := svc.base("Plan", options.Name)
// 	if context.Infrastructure == nil {
// 		return nil
// 	}

// 	commandArgs := make([]string, 0)
// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)
// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	if options.BuildDependencies {
// 		var dependencyError error
// 		stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			return nil
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return nil
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		if stack.VariableFile == "" || len(stack.Variables) > 0 {
// 			// Forcing isLocal
// 			if stack.Variables["isLocal"] == nil {
// 				stack.Variables["isLocal"] = true
// 			}

// 			path, err := svc.generateTerraformVariableFile(context, stack)
// 			if err != nil {
// 				notify.Error(err.Error())
// 				return nil
// 			}
// 			stack.VariableFile = path
// 		}

// 		if stackPath == "" || !helper.DirectoryExists(stackPath) {
// 			notify.Error("There was an error initiating the %s stack, err: path does not exists", stack.Name)
// 			return nil
// 		}

// 		commandArgs = append(commandArgs, fmt.Sprintf("-var-file=%s", stack.VariableFile))

// 		planOutputPath := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)
// 		if !helper.FileExists(planOutputPath) {
// 			notify.Hammer("Creating %s folder", planOutputPath)
// 			if !helper.CreateDirectory(planOutputPath, fs.ModePerm) {
// 				err := fmt.Errorf("error creating the %v folder", planOutputPath)
// 				notify.Error(err.Error())
// 				return nil
// 			}
// 		}

// 		planFileName := helper.JoinPath(planOutputPath, fmt.Sprintf("%s.plan", common.EncodeName(stack.Name)))
// 		if common.IsDebug() {
// 			notify.Debug("Plan File Path: %s", planFileName)
// 		}

// 		if stackResult, err := svc.wrapper.Plan(stackPath, planFileName, commandArgs...); err != nil {
// 			if err.Error() == MissingInitWhenApplying {
// 				notify.Warning("Stack %s was not initiated, trying to initiate it in this run", stack.Name)
// 				notify.Reset()
// 				stackOptions := &TerraformServiceOptions{
// 					Name:              options.Name,
// 					BuildDependencies: !options.BuildDependencies,
// 					RootFolder:        options.RootFolder,
// 					StackPath:         options.StackPath,
// 				}

// 				svc.InitiateStack(stackOptions)
// 				if notify.HasErrors() {
// 					return nil
// 				}

// 				if err := svc.wrapper.Apply(stackPath, planFileName, commandArgs...); err != nil {
// 					notify.Error(err.Error())
// 					return nil
// 				}
// 			} else {
// 				notify.Error(err.Error())
// 				return nil
// 			}
// 		} else {
// 			result.ChangeOps += stackResult.ChangeOps
// 			result.CreateOps += stackResult.CreateOps
// 			result.DeleteOps += stackResult.DeleteOps
// 			result.NoOps += stackResult.NoOps
// 			notify.Success("Planned successfully %s stack", stack.Name)
// 		}
// 	}

// 	return &result
// }

// func (svc *TerraformService) ApplyStack(options *TerraformServiceOptions) {
// 	config, context := svc.base("Apply", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	commandArgs := make([]string, 0)
// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)
// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	if options.BuildDependencies {
// 		var dependencyError error
// 		stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			return
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		var planChanges *PlanChanges
// 		var err error
// 		planOutputPath := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)
// 		planFileName := helper.JoinPath(planOutputPath, fmt.Sprintf("%s.plan", common.EncodeName(stack.Name)))
// 		if common.IsDebug() {
// 			notify.Debug("Plan File Path: %s", icons.IconFire, planFileName)
// 		}
// 		if !helper.FileExists(planFileName) {
// 			notify.InfoWithIcon(icons.IconBell, "Plan file for stack %s was not found on %s, starting plan", stack.Name, planOutputPath)
// 			planChanges = svc.PlanStack(options)
// 		} else {
// 			notify.InfoWithIcon(icons.IconThumbsUp, "Found plan file for stack %s on %s, using it to apply the stack", stack.Name, planOutputPath)
// 			planChanges, err = svc.wrapper.Show(stackPath, planFileName)
// 			if err != nil {
// 				return
// 			}
// 		}

// 		if common.IsDebug() {
// 			notify.Debug("Plan Changes %v", fmt.Sprintf("%v", planChanges))
// 		}

// 		shouldApply := false
// 		if planChanges == nil {
// 			//lint:ignore ST1005 directly used
// 			err := fmt.Errorf("There was an error reading the plan changes, err . plan changes are null, there should be some value")
// 			notify.Error(err.Error())
// 			return
// 		}

// 		icon := icons.IconThumbsUp
// 		hasAcceptedDestructive := false
// 		noChangesMessage := "Not applying stack as plan is up-to-date"
// 		if planChanges != nil && planChanges.HasChanges() {
// 			if planChanges.HasDestructiveChanges() {
// 				if helper.GetFlagSwitch("no-input", false) {
// 					shouldApply = true
// 				} else {
// 					notify.Warning("There are destructive actions in the plan, do you want to proceed?")
// 					reader := bufio.NewReader(os.Stdin)
// 					fmt.Printf("%sApprove? [yes/no]: ", icons.IconExclamationMark)
// 					approve, _ := reader.ReadString('\n')
// 					approve = strings.ReplaceAll(approve, "\r\n", "")
// 					approve = strings.ReplaceAll(approve, "\n", "")
// 					if strings.EqualFold(approve, "yes") {
// 						shouldApply = true
// 						hasAcceptedDestructive = true
// 					} else {
// 						hasAcceptedDestructive = false
// 						icon = icons.IconThumbDown
// 						noChangesMessage = "Not applying has it has destructive actions and were not approved"
// 					}
// 				}
// 			} else {
// 				hasAcceptedDestructive = true
// 				shouldApply = true
// 			}
// 		} else {
// 			hasAcceptedDestructive = true
// 		}

// 		if stack.LastApplied == nil {
// 			shouldApply = true
// 		}

// 		if shouldApply && hasAcceptedDestructive {
// 			if err := svc.wrapper.Apply(stackPath, planFileName, commandArgs...); err != nil {
// 				if err.Error() == MissingInitWhenApplying {
// 					notify.Warning(fmt.Sprintf("Stack %s was not initiated, trying to initiate it in this run", stack.Name))
// 					notify.Reset()
// 					initOptions := &TerraformServiceOptions{
// 						Name:              options.Name,
// 						BuildDependencies: !options.BuildDependencies,
// 						RootFolder:        options.RootFolder,
// 						StackPath:         options.StackPath,
// 					}

// 					svc.InitiateStack(initOptions)
// 					if notify.HasErrors() {
// 						return
// 					}
// 					if err := svc.wrapper.Apply(stackPath, planFileName, commandArgs...); err != nil {
// 						notify.Error(err.Error())
// 						return
// 					}
// 				} else {
// 					notify.Error(err.Error())
// 					return
// 				}
// 			}
// 		} else {
// 			notify.InfoWithIcon(icon, "%s", noChangesMessage)
// 		}

// 		if planFileName != "" {
// 			if err := helper.DeleteFile(planFileName); err != nil {
// 				notify.Error(err.Error())
// 				return
// 			}
// 		}

// 		if shouldApply || stack.LastApplied == nil {
// 			logger.Success("Applied successfully %s stack", stack.Name)
// 		}

// 		now := time.Now()
// 		stack.LastApplied = &now
// 		fragment := context.GetFragment(stack.GetSource())
// 		context.SaveFragment(fragment)
// 	}
// }

// func (svc *TerraformService) DestroyStack(options *TerraformServiceOptions) {
// 	config, context := svc.base("Destroy", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	commandArgs := make([]string, 0)
// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)
// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	if context.HasTags() || options.BuildDependencies {
// 		var err error
// 		stacks, err = config.GetInfrastructureDependencies(stacks)
// 		if err != nil {
// 			notify.FromError(err, "Building the infrastructure dependencies")
// 			return
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	dependency_tree.ReverseDependency(stacks)
// 	stackNames := ""

// 	for i, stack := range stacks {
// 		if i < len(stacks) && i > 0 {
// 			stackNames += ", "
// 		}
// 		stackNames += stack.Name
// 	}

// 	shouldApply := false
// 	notify.InfoWithIcon(icons.IconBomb, "ATTENTION: This will destroy the following infrastructure: %s\nDo you want to proceed?", stackNames)
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Printf("%sApprove? [yes/no]: ", icons.IconExclamationMark)
// 	approve, _ := reader.ReadString('\n')
// 	approve = strings.ReplaceAll(approve, "\r\n", "")
// 	approve = strings.ReplaceAll(approve, "\n", "")
// 	if strings.EqualFold(approve, "yes") {
// 		shouldApply = true
// 	}

// 	if !shouldApply {
// 		notify.InfoWithIcon(icons.IconThumbDown, "Destroy of %s stacks was canceled by user", stackNames)
// 		return
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		if stack.VariableFile == "" || len(stack.Variables) > 0 {
// 			// Forcing isLocal
// 			if stack.Variables["isLocal"] == nil {
// 				stack.Variables["isLocal"] = true
// 			}

// 			path, err := svc.generateTerraformVariableFile(context, stack)
// 			if err != nil {
// 				notify.Error(err.Error())
// 				return
// 			}
// 			stack.VariableFile = path
// 		}

// 		if stackPath == "" || !helper.DirectoryExists(stackPath) {
// 			//lint:ignore ST1005 directly used
// 			err := fmt.Errorf("There was an error destroying the %s stack, err. path does not exists", stack.Name)
// 			notify.Error(err.Error())
// 			return
// 		}

// 		commandArgs = append(commandArgs, fmt.Sprintf("-var-file=%s", stack.VariableFile))

// 		backupPath := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)
// 		if !helper.FileExists(backupPath) {
// 			notify.Hammer("Creating %s folder", backupPath)
// 			if !helper.CreateDirectory(backupPath, fs.ModePerm) {
// 				err := fmt.Errorf("error creating the %v folder", backupPath)
// 				notify.Error(err.Error())
// 				return
// 			}
// 		}

// 		backupFileName := helper.JoinPath(backupPath, fmt.Sprintf("%s.backup", common.EncodeName(stack.Name)))
// 		if common.IsDebug() {
// 			notify.Debug("Backup File Path: %s", backupFileName)
// 		}

// 		if err := svc.wrapper.Destroy(stackPath, backupFileName, commandArgs...); err != nil {
// 			notify.Error(err.Error())
// 			return
// 		} else {
// 			notify.InfoWithIcon(icons.IconThumbsUp, "Destroyed successfully %s stack", stack.Name)
// 		}
// 	}

// 	// removing content from the infrastructure folder
// 	removeDir := helper.JoinPath(context.Configuration.OutputPath, common.INFRASTRUCTURE_PATH)
// 	if err := helper.DeleteAllFiles(removeDir); err != nil {
// 		notify.Warning("Could not remove the files on the infrastructure folder")
// 	} else {
// 		notify.InfoWithIcon(icons.IconThumbsUp, "Cleaned successfully the infrastructure folder")
// 	}
// }

// func (svc *TerraformService) OutputStack(options *TerraformServiceOptions) {
// 	env := environment.GetInstance()
// 	config, context := svc.base("Output", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	stacks := context.GetInfrastructureStacks(options.Name, !options.BuildDependencies)

// 	if options.BuildDependencies {
// 		var dependencyError error
// 		stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			return
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Stacks dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Stack %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		outputs, err := svc.wrapper.Output(stackPath)
// 		if err != nil {
// 			notify.Error(err.Error())
// 			return
// 		}

// 		for key, value := range outputs {
// 			if context.EnvironmentVariables != nil && context.EnvironmentVariables.Terraform == nil {
// 				context.EnvironmentVariables.Terraform = make(map[string]interface{}, 0)
// 			}

// 			terraformKey := fmt.Sprintf("%s.%s", strings.ToLower(common.EncodeName(stack.Name)), strings.ToLower(key))
// 			if common.IsDebug() {
// 				notify.Debug("Setting the terraform variable %s to the global environments %s", key, terraformKey)
// 			}

// 			if value.Value != "" {
// 				env.Add("terraform", terraformKey, value.Value)
// 				if context.EnvironmentVariables != nil && context.EnvironmentVariables.Terraform != nil {
// 					context.EnvironmentVariables.Terraform[terraformKey] = value.Value
// 				}
// 			}
// 		}

// 		context.SaveEnvironmentVariables()
// 		if common.IsDebug() {
// 			notify.Debug("Config Global Environment Variables:\n  %s", fmt.Sprintf("%v", context.EnvironmentVariables))
// 		}
// 	}
// }

// func (svc *TerraformService) GraphStack(options *TerraformServiceOptions) {
// 	config, context := svc.base("Generating dependency graph", options.Name)
// 	if context.Infrastructure == nil {
// 		return
// 	}

// 	notify.InfoWithIcon(icons.IconMagnifyingGlass, "This can take a while...")

// 	stacks := context.GetInfrastructureStacks(options.Name, true)
// 	if len(stacks) == 0 {
// 		notify.Warning("No infrastructure stacks found")
// 	}

// 	for _, stack := range stacks {
// 		stackPath, pathError := svc.getPath(stack, options)
// 		if pathError != nil {
// 			notify.FromError(pathError, "Could not get the path")
// 			return
// 		}
// 		notify.Debug("Using Path: %s", stackPath)

// 		notify.Debug("Stack: %s", stack.Name)
// 		graph, err := svc.wrapper.DependencyGraph(stackPath)
// 		if err != nil {
// 			return
// 		}

// 		if options.StdOutput {
// 			o, err := yaml.Marshal(graph)
// 			if err == nil {
// 				notify.Info(string(o))
// 			}
// 		}
// 		if err != nil {
// 			notify.Error(err.Error())
// 		}

// 		tfstate := make(map[string]interface{})
// 		if stack.Variables["tfstate"] != nil {
// 			for k, v := range stack.Variables["tfstate"].(map[string]interface{}) {
// 				tfstate[k] = v
// 			}
// 		}

// 		if stack.Backend.ContainerName != "" && (tfstate["container_name"] == "" || tfstate["container_name"] == nil) {
// 			tfstate["container_name"] = stack.Backend.ContainerName
// 		}
// 		if stack.Backend.ResourceGroupName != "" && (tfstate["resource_group_name"] == "" || tfstate["resource_group_name"] == nil) {
// 			tfstate["resource_group_name"] = stack.Backend.ResourceGroupName
// 		}
// 		if stack.Backend.StorageAccountName != "" && (tfstate["storage_account_name"] == "" || tfstate["storage_account_name"] == nil) {
// 			tfstate["storage_account_name"] = stack.Backend.StorageAccountName
// 		}

// 		fragment := context.GetFragment(stack.GetSource())
// 		if fragment != nil {
// 			fragmentStack := config.GetFragmentInfrastructureStack(fragment, stack.Name)
// 			if fragmentStack != nil {
// 				states := graph.GetStateQuery()
// 				for _, s := range states {
// 					notify.Debug("State Label: %s", s.Label)
// 					notify.Debug("State Name: %s", s.Name)
// 					notify.Debug("Adding dependency on state %s to %s", s.Name, stack.Name)
// 					// Trying to find the backend key in our configuration
// 					backendKey := s.Name
// 					notify.Debug("State Backend Key: %s", backendKey)
// 					backendStack := context.Infrastructure.GetStackByBackend(backendKey)
// 					if backendStack != nil {
// 						stack.AddDependency(backendStack.Name)
// 						stack.AddRequiredState(backendStack.Name)
// 						backendStack.AddRequiredBy(stack.Name)
// 						fragment := context.GetFragment(backendStack.GetSource())
// 						if fragment != nil {
// 							if err := fragment.SaveFragment(fragment); err != nil {
// 								notify.Warning(err.Error())
// 							}
// 						} else {
// 							notify.Warning("Could not find %s dependency fragment file to update required by", backendStack.Name)
// 						}

// 						key := strings.ReplaceAll(backendStack.Name, "_stack", "_key")
// 						key = strings.ReplaceAll(key, "-stack", "_key")
// 						kv := backendStack.Backend.StateFileName
// 						if !strings.HasSuffix(kv, ".tfstate") {
// 							kv = fmt.Sprintf("%s.tfstate", kv)
// 						}
// 						tfstate[key] = kv

// 					} else {
// 						notify.Critical("Could not find state dependency %s in our configuration file", backendKey)
// 					}
// 				}
// 			}

// 			notify.Debug("tfstate variable not found, trying to generating it")
// 			stack.Variables["tfstate"] = tfstate

// 			err := fragment.SaveFragment(fragment)
// 			if err != nil {
// 				notify.Error(err.Error())
// 				return
// 			}
// 		}
// 	}
// }

// func (svc *TerraformService) CanClone(stack *infrastructure_component.InfrastructureStack) bool {
// 	if stack == nil {
// 		notify.Debug("Stack is nil, ignoring")
// 		return false
// 	}

// 	if stack.Location != nil && stack.Location.Path != "" && stack.Location.RootFolder != "" {
// 		if stack.Repository != nil && !stack.Repository.Enabled {
// 			currentPath := helper.JoinPath(stack.Location.RootFolder, stack.Location.Path)
// 			notify.Debug("Stack has a path defined: %s, ignoring", currentPath)
// 			return false
// 		} else {
// 			notify.Debug("Stack has a path defined but repo is enabled")
// 		}
// 	}

// 	if stack.Repository == nil {
// 		notify.Debug("Stack has no repository defined, ignoring")
// 		return false
// 	}

// 	if !stack.Repository.Enabled {
// 		notify.Debug("Stack repository was not enabled, ignoring")
// 		return false
// 	}

// 	if stack.Repository.Url == "" {
// 		notify.Debug("Stack has no repository url defined, ignoring")
// 		return false
// 	}

// 	return true
// }

// func (svc *TerraformService) clone(stack *infrastructure_component.InfrastructureStack) (string, error) {
// 	env := environment.GetInstance()

// 	destination := env.Replace(stack.Repository.Destination)
// 	if helper.DirectoryExists(destination) {
// 		notify.Debug("Destination folder %s already exists, ignoring", destination)
// 		return destination, nil
// 	}

// 	cleanRepo := helper.GetFlagSwitch("clean-repo", false)

// 	if svc.CanClone(stack) {
// 		git := git.Get()

// 		if stack.Repository.Credentials != nil {
// 			mappers.DecodeGitCredentials(stack.Repository.Credentials)
// 			if err := git.CloneWithCredentials(stack.Repository.Url, destination, stack.Repository.Credentials, cleanRepo); err != nil {
// 				return "", err
// 			}
// 		} else {
// 			if err := git.Clone(stack.Repository.Url, destination, false); err != nil {
// 				return "", err
// 			}
// 		}

// 		return destination, nil
// 	} else {
// 		return "", nil
// 	}
// }

// func (svc *TerraformService) getPath(stack *infrastructure_component.InfrastructureStack, options *TerraformServiceOptions) (string, error) {
// 	env := environment.GetInstance()
// 	returnPath := ""

// 	if options == nil {
// 		err := fmt.Errorf("options cannot be nil")
// 		return "", err
// 	}

// 	if options.RootFolder != "" {
// 		returnPath = helper.JoinPath(env.Replace(options.RootFolder), env.Replace(stack.Location.Path))
// 	}

// 	if stack.Location != nil && stack.Location.RootFolder != "" {
// 		returnPath = helper.JoinPath(env.Replace(stack.Location.RootFolder), env.Replace(stack.Location.Path))
// 	}

// 	if stack.Repository != nil && stack.Repository.Enabled {
// 		returnPath = ""
// 	}

// 	if returnPath == "" {
// 		if stack.Repository == nil {
// 			return "", nil
// 		}
// 		clonedPath, err := svc.clone(stack)
// 		if err != nil {
// 			return "", err
// 		}
// 		if clonedPath != "" {
// 			return helper.JoinPath(clonedPath, env.Replace(stack.Location.Path)), nil
// 		} else {
// 			return helper.JoinPath(env.Replace(stack.Location.RootFolder), env.Replace(stack.Location.Path)), nil
// 		}
// 	}

// 	returnPath = strings.Trim(returnPath, "\\")
// 	notify.Debug("Got the return path of %s for infrastructure service", returnPath)

// 	return returnPath, nil
// }
