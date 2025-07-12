package context

import (
	"strings"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/context/infrastructure_component"
	"github.com/cjlapao/locally-cli/internal/context/nuget_package_component"
	"github.com/cjlapao/locally-cli/internal/context/service_component"
	"github.com/cjlapao/locally-cli/internal/icons"

	"github.com/cjlapao/common-go/helper"
)

func (ctx *Context) HasTags() bool {
	tags := helper.GetFlagArrayValue("tag")
	return len(tags) > 0
}

func (ctx *Context) GetSpaServicesByTags() []*service_component.SpaService {
	tags := helper.GetFlagArrayValue("tag")

	if len(tags) == 0 {
		return ctx.SpaServices
	}

	result := make([]*service_component.SpaService, 0)
	for _, tag := range tags {
		notify.Debug("Finding spa services matching command line tag %s", tag)
		for _, spaService := range ctx.SpaServices {
			notify.Debug("  Matching the spa service %s tags", spaService.Name)
			for _, serviceTag := range spaService.Tags {
				notify.Debug("  Matching the spa service %s tag %s against command line tag %s", spaService.Name, serviceTag, tag)
				if strings.EqualFold(tag, serviceTag) {
					exists := false
					notify.Debug("  Matched %s spa service tag against %s command line tag in backend service %s tag %s against command line tag %s", tag, serviceTag, spaService.Name)
					for _, addedSpaService := range result {
						if strings.EqualFold(addedSpaService.Name, spaService.Name) {
							exists = true
							break
						}
					}

					if !exists {
						notify.Debug("  Spa service %s was not found on the list, appending it", spaService.Name)
						result = append(result, spaService)
					}
				}
			}
		}
	}
	return result
}

func (ctx *Context) GetBackendServicesByTags() []*service_component.BackendService {
	tags := helper.GetFlagArrayValue("tag")

	if len(tags) == 0 {
		return ctx.BackendServices
	}

	result := make([]*service_component.BackendService, 0)
	for _, tag := range tags {
		notify.Debug("Finding services matching command line tag %s", tag)
		for _, backendService := range ctx.BackendServices {
			notify.Debug("  Matching the backend service %s tags", backendService.Name)
			for _, backendServiceTag := range backendService.Tags {
				notify.Debug("  Matching the backend service %s tag %s against command line tag %s", backendService.Name, backendServiceTag, tag)
				if strings.EqualFold(tag, backendServiceTag) {
					exists := false
					notify.Debug("  Matched %s backend tag against %s command line tag in backend service %s tag %s against command line tag %s", tag, backendServiceTag, backendService.Name)

					for _, addedBackendService := range result {
						if strings.EqualFold(addedBackendService.Name, backendService.Name) {
							exists = true
							break
						}
					}

					if !exists {
						notify.Debug("  Backend %s was not found on the list, appending it", icons.IconFire, backendService.Name)
						result = append(result, backendService)
					}
				}
			}
		}
	}
	return result
}

func (ctx *Context) GetNugetPackagesByTags() []*nuget_package_component.NugetPackage {
	tags := helper.GetFlagArrayValue("tag")

	if ctx.NugetPackages == nil {
		return make([]*nuget_package_component.NugetPackage, 0)
	}

	if len(tags) == 0 {
		return ctx.NugetPackages.Packages
	}

	result := make([]*nuget_package_component.NugetPackage, 0)
	for _, tag := range tags {
		for _, nugetPackage := range ctx.NugetPackages.Packages {
			for _, nugetPackageTag := range nugetPackage.Tags {
				if strings.EqualFold(tag, nugetPackageTag) {
					exists := false
					for _, addedNugetPackage := range result {
						if strings.EqualFold(addedNugetPackage.Name, nugetPackage.Name) {
							exists = true
							break
						}
					}

					if !exists {
						result = append(result, nugetPackage)
					}
				}
			}
		}
	}

	return result
}

func (ctx *Context) GetInfrastructureStacksByTags() []*infrastructure_component.InfrastructureStack {
	tags := helper.GetFlagArrayValue("tag")

	if ctx.Infrastructure == nil {
		return make([]*infrastructure_component.InfrastructureStack, 0)
	}

	if len(tags) == 0 {
		return ctx.Infrastructure.Stacks
	}

	result := make([]*infrastructure_component.InfrastructureStack, 0)
	for _, tag := range tags {
		for _, stack := range ctx.Infrastructure.Stacks {
			for _, stackTag := range stack.Tags {
				if strings.EqualFold(tag, stackTag) {
					exists := false
					for _, addedStack := range result {
						if strings.EqualFold(addedStack.Name, stack.Name) {
							exists = true
							break
						}
					}

					if !exists {
						result = append(result, stack)
					}
				}
			}
		}
	}

	return result
}

func (ctx *Context) GetInfrastructureStacks(name string, ignoreTags bool) []*infrastructure_component.InfrastructureStack {
	var stacks []*infrastructure_component.InfrastructureStack
	if ignoreTags {
		notify.Debug("Ignoring flags, getting one by one")
		stacks = ctx.Infrastructure.Stacks
	} else {
		stacks = ctx.GetInfrastructureStacksByTags()
	}

	if helper.GetFlagSwitch("all", false) || ctx.HasTags() && !ignoreTags {
		return stacks
	} else {
		if name == "" {
			return make([]*infrastructure_component.InfrastructureStack, 0)
		}

		filteredStacks := make([]*infrastructure_component.InfrastructureStack, 0)
		for _, stack := range stacks {
			if strings.EqualFold(common.EncodeName(stack.Name), name) {
				filteredStacks = append(filteredStacks, stack)
			}
		}

		return filteredStacks
	}
}
