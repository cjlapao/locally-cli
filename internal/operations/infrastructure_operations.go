package operations

// import (
// 	"os"

// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/help"
// 	"github.com/cjlapao/locally-cli/internal/icons"
// 	"github.com/cjlapao/locally-cli/internal/infrastructure"
// 	"github.com/cjlapao/locally-cli/internal/notifications"

// 	"github.com/cjlapao/common-go/helper"
// )

// func InfrastructureOperations(subCommand, stack string, options *infrastructure.TerraformServiceOptions) {
// 	config := configuration.Get()
// 	context := config.GetCurrentContext()
// 	terraformSvc := infrastructure.Get()
// 	notify := notifications.Get()

// 	if err := terraformSvc.InitBackendResources(); err != nil {
// 		notify.FromError(err, "Missing backend configuration")
// 		return
// 	}

// 	terraformSvc.CheckForTerraform(false)
// 	if err := terraformSvc.CheckForCredentials(); err != nil {
// 		notify.FromError(err, "Could not create required service principal for running %s infrastructure command on stack %s", subCommand, stack)
// 		return
// 	}

// 	if subCommand == "" && helper.GetFlagSwitch("help", false) {
// 		help.ShowHelpForInfrastructureCommand()
// 		os.Exit(0)
// 	}

// 	switch subCommand {
// 	case "init-backend":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if err := terraformSvc.InitBackendResources(); err != nil {
// 			notify.Error(err.Error())
// 		}

// 	case "init":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}
// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.InitiateStack(options)

// 	case "validate":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		notify.Debug("Validating with the options: %v", options)
// 		terraformSvc.ValidateStack(options)

// 	case "plan":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := false
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.PlanStack(options)

// 	case "apply":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := false
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.ApplyStack(options)
// 		if !notify.HasErrors() {
// 			options.BuildDependencies = false
// 			terraformSvc.OutputStack(options)
// 		}

// 	case "destroy":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: false,
// 			}
// 		}

// 		terraformSvc.DestroyStack(options)

// 	case "graph":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: false,
// 				StdOutput:         true,
// 			}
// 		}

// 		terraformSvc.GraphStack(options)

// 	case "output":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.OutputStack(options)

// 	case "up":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureUpCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureUpCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		stacks := context.GetInfrastructureStacks(stack, !buildDependencies)
// 		if len(stacks) == 0 {
// 			notify.Warning("No infrastructure stacks found")
// 		}

// 		if buildDependencies {
// 			var dependencyError error
// 			stacks, dependencyError = config.GetInfrastructureDependencies(stacks)
// 			if dependencyError != nil {
// 				notify.FromError(dependencyError, "Building the infrastructure dependencies")
// 			} else {
// 				if stack == "" {
// 					notify.Hammer("Stacks dependencies were built successfully")
// 				} else {
// 					notify.Hammer("Stack %s dependencies were built successfully", stack)
// 				}
// 			}
// 		}

// 		if buildDependencies {
// 			notify.Info("Starting to bring up the infrastructure following the dependency order of:")
// 			for _, dependentStack := range stacks {
// 				notify.InfoWithIcon(icons.IconFlag, "%s stack", dependentStack.Name)
// 			}
// 		}

// 		// We need to init and validate all dependent stacks before we commit to plan and apply
// 		for _, dependentStack := range stacks {
// 			dependencyOptions := &infrastructure.TerraformServiceOptions{
// 				Name:              dependentStack.Name,
// 				BuildDependencies: false,
// 			}
// 			terraformSvc.InitiateStack(dependencyOptions)
// 			if !notify.HasErrors() {
// 				terraformSvc.ValidateStack(dependencyOptions)
// 			}
// 		}

// 		// In the init we need to plan/apply each stack individual as the dependency between them
// 		// might require it
// 		for _, dependentStack := range stacks {
// 			if !notify.HasErrors() {
// 				dependencyOptions := &infrastructure.TerraformServiceOptions{
// 					Name:              dependentStack.Name,
// 					BuildDependencies: false,
// 				}
// 				notify.Reset()
// 				terraformSvc.PlanStack(dependencyOptions)
// 				notify.Info("Finished planning lets apply %s", dependentStack.Name)
// 				if !notify.HasErrors() {
// 					notify.Info("Applying %s", dependentStack.Name)
// 					terraformSvc.ApplyStack(dependencyOptions)
// 					if !notify.HasErrors() {
// 						terraformSvc.OutputStack(dependencyOptions)
// 					}
// 				}
// 			}
// 		}

// 	case "down":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.DestroyStack(options)

// 	case "refresh":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		if stack == "" && !context.HasTags() {
// 			help.ShowHelpForInfrastructureCommand()
// 			os.Exit(0)
// 		}

// 		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
// 		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
// 		if context.HasTags() {
// 			buildDependencies = true
// 		}
// 		if options != nil && options.BuildDependencies {
// 			buildDependencies = true
// 		}

// 		if options == nil {
// 			options = &infrastructure.TerraformServiceOptions{
// 				Name:              stack,
// 				BuildDependencies: buildDependencies,
// 			}
// 		}

// 		terraformSvc.InitiateStack(options)
// 		if !notify.HasErrors() {
// 			terraformSvc.OutputStack(options)
// 		}
// 	default:
// 		help.ShowHelpForInfrastructureCommand()
// 		os.Exit(0)
// 	}
// }
