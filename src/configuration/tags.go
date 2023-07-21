package configuration

import (
	"github.com/cjlapao/locally-cli/icons"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

func (svc *ConfigService) HasTags() bool {
	tags := helper.GetFlagArrayValue("tag")
	return len(tags) > 0
}

func (svc *ConfigService) GetSpaServicesByTags() []*SpaService {
	tags := helper.GetFlagArrayValue("tag")
	config := Get()

	if len(tags) == 0 {
		return config.GetCurrentContext().SpaServices
	}

	result := make([]*SpaService, 0)
	for _, tag := range tags {
		notify.Debug("Finding spa services matching command line tag %s", tag)
		for _, spaService := range config.GetCurrentContext().SpaServices {
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

func (svc *ConfigService) GetBackendServicesByTags() []*BackendService {
	tags := helper.GetFlagArrayValue("tag")
	config := Get()

	if len(tags) == 0 {
		return config.GetCurrentContext().BackendServices
	}

	result := make([]*BackendService, 0)
	for _, tag := range tags {
		notify.Debug("Finding services matching command line tag %s", tag)
		for _, backendService := range config.GetCurrentContext().BackendServices {
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

func (svc *ConfigService) GetNugetPackagesByTags() []*NugetPackage {
	tags := helper.GetFlagArrayValue("tag")
	config := Get()

	if config.GetCurrentContext().NugetPackages == nil {
		return make([]*NugetPackage, 0)
	}

	if len(tags) == 0 {
		return config.GetCurrentContext().NugetPackages.Packages
	}

	result := make([]*NugetPackage, 0)
	for _, tag := range tags {
		for _, nugetPackage := range config.GetCurrentContext().NugetPackages.Packages {
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

func (svc *ConfigService) GetInfrastructureStacksByTags() []*InfrastructureStack {
	tags := helper.GetFlagArrayValue("tag")
	config := Get()

	if config.GetCurrentContext().Infrastructure == nil {
		return make([]*InfrastructureStack, 0)
	}

	if len(tags) == 0 {
		return config.GetCurrentContext().Infrastructure.Stacks
	}

	result := make([]*InfrastructureStack, 0)
	for _, tag := range tags {
		for _, stack := range config.GetCurrentContext().Infrastructure.Stacks {
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

func (svc *ConfigService) GetInfrastructureStacks(name string, ignoreTags bool) []*InfrastructureStack {
	var stacks []*InfrastructureStack
	if ignoreTags {
		notify.Debug("Ignoring flags, getting one by one")
		stacks = svc.GetCurrentContext().Infrastructure.Stacks
	} else {
		stacks = svc.GetInfrastructureStacksByTags()
	}

	if helper.GetFlagSwitch("all", false) || svc.HasTags() && !ignoreTags {
		return stacks
	} else {
		if name == "" {
			return make([]*InfrastructureStack, 0)
		}

		filteredStacks := make([]*InfrastructureStack, 0)
		for _, stack := range stacks {
			if strings.EqualFold(EncodeName(stack.Name), name) {
				filteredStacks = append(filteredStacks, stack)
			}
		}

		return filteredStacks
	}
}
