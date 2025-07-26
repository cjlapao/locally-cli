// Package dependency_tree provides functionality to build and analyze dependency graphs.
// It includes functions to detect cycles, self-dependencies, and topological sorting.
// The package also includes utility functions for shifting dependencies and updating dependency trees.
// The main entry point is BuildDependencyTree, which takes a list of services and returns a sorted list
// based on their dependencies.
//
// The package uses Kahn's algorithm for topological sorting, which is a linear-time algorithm for
// directed acyclic graphs (DAGs). It ensures that dependencies are resolved in a valid order,
// avoiding cycles and self-dependencies.
package dependency_tree

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/context"
	"github.com/cjlapao/locally-cli/internal/notifications"
	"github.com/cjlapao/locally-cli/pkg/interfaces"
)

var notify = notifications.Get()

type TreeItem struct {
	CurrentIndex int
	Highest      int
	Lowest       int
}

func ReverseDependency[T interfaces.LocallyService](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func BuildDependencyTree[T interfaces.LocallyService](values []T) ([]T, error) {
	// Detect cycles/self-dependencies before proceeding
	if err := detectCycles(values); err != nil {
		return nil, err
	}

	if common.IsDebug() {
		notify.Debug("Dependency Tree Before:")
		for idx, service := range values {
			notify.Debug("[%s] %s", strconv.Itoa(idx), service.GetName())
		}
	}

	// Topological sort (Kahn's algorithm)
	nameToService := make(map[string]T)
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	for _, svc := range values {
		name := strings.ToLower(svc.GetName())
		nameToService[name] = svc
		inDegree[name] = 0
	}
	for _, svc := range values {
		name := strings.ToLower(svc.GetName())
		for _, dep := range svc.GetDependencies() {
			depName := strings.ToLower(dep)
			if _, ok := nameToService[depName]; ok {
				graph[depName] = append(graph[depName], name)
				inDegree[name]++
			}
		}
	}

	// Collect nodes with in-degree 0
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	var sorted []T
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, nameToService[current])
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(values) {
		return nil, fmt.Errorf("dependency tree could not be resolved (possible cycle or missing node)")
	}

	if common.IsDebug() {
		notify.Debug("Dependency Tree After:")
		for idx, service := range sorted {
			notify.Debug("[%s] %s", strconv.Itoa(idx), service.GetName())
		}
	}

	return sorted, nil
}

func BuildDependencyGraph[T interfaces.LocallyService](ctx *context.Context, values []T, persist bool) error {
	// notify.InfoWithIcon(icons.IconMagnifyingGlass, "This can take a while...")
	if _, err := BuildDependencyTree(values); err != nil {
		notify.Error(err.Error())
		return err
	}

	for _, service := range values {
		notify.Debug("Service: %s from source %s", service.GetName(), service.GetSource())
		fragment := ctx.GetFragment(service.GetSource())
		if fragment != nil {
			for _, dependsOn := range service.GetDependencies() {
				notify.Debug("Service %s depends on %s", service.GetName(), dependsOn)
				requiredSvc := ctx.GetRegisteredService(dependsOn)
				if requiredSvc == nil {
					return fmt.Errorf("could not find dependent service %s", dependsOn)
				}
				requiredSvc.AddRequiredBy(service.GetName())
				dependsOnFragment := ctx.GetFragment(requiredSvc.GetSource())
				if dependsOnFragment == nil {
					return fmt.Errorf("could not find fragment for %s on source %s", dependsOn, requiredSvc.GetSource())
				}

				// Saving the changes
				if persist {
					if err := dependsOnFragment.SaveFragment(dependsOnFragment); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func buildRightDependency[T interfaces.LocallyService](s []T) ([]T, error) {
	idx := 0
	for {
		needsShifting := false
		for svcIndex, service := range s {
			dependencies := service.GetDependencies()
			for _, dependency := range dependencies {
				dependencyIndex := getIndex(s, dependency)
				if dependencyIndex < 0 {
					err := fmt.Errorf("dependency on %s of service %s was not found in the context configuration", dependency, service.GetName())
					return nil, err
				}

				if svcIndex < dependencyIndex {
					needsShifting = true
					if common.IsDebug() && common.IsVerbose() {
						notify.Debug("Shifting %s on index %s to index %s", service.GetName(), strconv.Itoa(svcIndex), strconv.Itoa(dependencyIndex))
					}

					shiftRight(s, dependencyIndex-1)
				}
			}

			if needsShifting {
				break
			}
		}

		idx += 1
		if !needsShifting || idx > 1000 {
			if idx == 1000 {
				err := fmt.Errorf("something went wrong and we shifted more than 1000 items")
				return nil, err
			}

			break
		}
	}

	return s, nil
}

func buildLeftDependency[T interfaces.LocallyService](s []T) ([]T, error) {
	idx := 0
	for {
		shiftHappened := false
		// For each loop after a shift we need to re-update our tree so we know of the position shifting
		tree := updateTree(s)
		for _, service := range s {
			item := tree[service.GetName()]
			highestIndex := -1
			changeTo := ""
			change := true

			if item.Highest == -1 || item.Lowest == -1 {
				continue
			}

			if item.CurrentIndex < item.Lowest || item.CurrentIndex-1 > item.Highest {
				for k, kv := range tree {
					if kv.Highest == item.Highest {
						if kv.CurrentIndex > highestIndex && kv.CurrentIndex != item.CurrentIndex {
							highestIndex = kv.CurrentIndex
							changeTo = k
						}

						if kv.CurrentIndex == item.CurrentIndex {
							if item.CurrentIndex-1 == highestIndex {
								change = false
								break
							}
						}
					}
				}

				if change && highestIndex != -1 && item.CurrentIndex > highestIndex && item.CurrentIndex != highestIndex+1 {
					if common.IsDebug() && common.IsVerbose() {
						notify.Debug("[%s] %s should move to %s after %s", fmt.Sprintf("%d", item.CurrentIndex), service.GetName(), fmt.Sprintf("%d", highestIndex+1), changeTo)
					}
					shiftTo(s, item.CurrentIndex, highestIndex+1)
					shiftHappened = true
					break
				}
			}
		}

		idx += 1
		if !shiftHappened || idx > 1000 {
			if idx == 1000 {
				err := fmt.Errorf("something went wrong and we shifted more than 1000 items")
				return nil, err
			}
			break
		}
	}

	return s, nil
}

func shiftRight[T interfaces.LocallyService](s []T, index int) []T {
	if len(s) > 1 {
		for i, item := range s {
			if i == len(s)-1 {
				break
			}

			if i <= index {
				s[i] = s[i+1]
				s[i+1] = item
			}

			if i == index {
				break
			}
		}
	}

	return s
}

func shiftTo[T interfaces.LocallyService](s []T, from, to int) []T {
	if len(s) > 1 {
		// not inside the range of the array
		if from > len(s) || from == -1 || to > len(s) || to == -1 {
			return s
		}

		// Deciding the direction of the shift
		pos := from - to
		if pos < 0 {
			if common.IsDebug() && common.IsVerbose() {
				notify.Debug("Shifting forwards from %s to %s", strconv.Itoa(from), strconv.Itoa(to))
			}
			for {
				if from == to {
					break
				}
				item := s[from]
				s[from] = s[from+1]
				s[from+1] = item

				from += 1
			}
		}
		if pos > 0 {
			if common.IsDebug() && common.IsVerbose() {
				notify.Debug("Shifting backwards from %s to %s", strconv.Itoa(from), strconv.Itoa(to))
			}
			for {
				if from == to {
					break
				}
				item := s[from]
				s[from] = s[from-1]
				s[from-1] = item

				from -= 1
			}
		}
	}

	return s
}

func getIndex[T interfaces.LocallyService](s []T, name string) int {
	for idx, svc := range s {
		if strings.EqualFold(svc.GetName(), name) {
			return idx
		}
	}

	return -1
}

func updateTree[T interfaces.LocallyService](s []T) map[string]TreeItem {
	// Checking if there is any more order shifting
	tree := make(map[string]TreeItem)
	for i, service := range s {
		item := TreeItem{
			CurrentIndex: i,
			Highest:      -1,
			Lowest:       -1,
		}

		for _, dependency := range service.GetDependencies() {
			dependencyIndex := getIndex(s, dependency)
			if dependencyIndex > item.Highest {
				item.Highest = dependencyIndex
				if item.Lowest == -1 {
					item.Lowest = dependencyIndex
				}
			}
			if dependencyIndex < item.Highest && dependencyIndex > item.Lowest {
				item.Lowest = dependencyIndex
			}
		}

		if common.IsDebug() && common.IsVerbose() {
			notify.Debug("%s [%s] %s -> HighestDependency: %s | LowestDependency: %s", fmt.Sprintf("%d", i), service.GetName(), fmt.Sprintf("%d", item.Highest), fmt.Sprintf("%d", item.Lowest))
		}
		tree[service.GetName()] = item
	}

	return tree
}

// detectCycles returns an error if there is a cycle or self-dependency in the dependency graph
func detectCycles[T interfaces.LocallyService](services []T) error {
	// Build a map for quick lookup
	serviceMap := make(map[string]T)
	for _, svc := range services {
		serviceMap[strings.ToLower(svc.GetName())] = svc
	}

	visited := make(map[string]bool)
	stack := make(map[string]bool)

	var visit func(string) error
	visit = func(name string) error {
		lname := strings.ToLower(name)
		if stack[lname] {
			return fmt.Errorf("circular dependency detected involving service '%s'", name)
		}
		if visited[lname] {
			return nil
		}
		visited[lname] = true
		stack[lname] = true
		svc, ok := serviceMap[lname]
		if !ok {
			return fmt.Errorf("service '%s' not found in dependency graph", name)
		}
		for _, dep := range svc.GetDependencies() {
			if strings.EqualFold(dep, name) {
				return fmt.Errorf("self-dependency detected for service '%s'", name)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		stack[lname] = false
		return nil
	}

	for _, svc := range services {
		if err := visit(svc.GetName()); err != nil {
			return err
		}
	}
	return nil
}

// PrintDependencyGraph prints the dependency graph in a readable format for debugging
func PrintDependencyGraph[T interfaces.LocallyService](services []T) string {
	var b strings.Builder
	serviceMap := make(map[string]T)
	for _, svc := range services {
		serviceMap[svc.GetName()] = svc
	}
	b.WriteString("Dependency Graph:\n")
	for _, svc := range services {
		b.WriteString(fmt.Sprintf("- %s\n", svc.GetName()))
		deps := svc.GetDependencies()
		if len(deps) > 0 {
			b.WriteString("    depends on: ")
			for i, dep := range deps {
				b.WriteString(dep)
				if i != len(deps)-1 {
					b.WriteString(", ")
				}
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
