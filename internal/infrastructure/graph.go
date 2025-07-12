package infrastructure

import (
	"fmt"
	"strconv"
	"strings"
)

type TerraformGraph struct {
	Compound bool
	NewRank  bool
	Items    []*TerraformGraphItem
}

func (g *TerraformGraph) GetStateQuery() []*TerraformGraphItem {
	if g.Items == nil {
		return make([]*TerraformGraphItem, 0)
	}

	result := make([]*TerraformGraphItem, 0)
	for _, i := range g.Items {
		if strings.EqualFold(i.Type, "data") && strings.EqualFold(i.Resource, "terraform_remote_state") {
			notify.Debug("Found state query %s", i.Label)
			result = append(result, i)
		}
	}

	return result
}

type TerraformGraphItem struct {
	Level      string
	Module     string
	Resource   string
	Type       string
	Name       string
	Label      string
	DependsOn  []string
	RequiredBy []string
}

func (i *TerraformGraphItem) GetLabel() string {
	if i.Label != "" {
		return i.Label
	}

	label := i.Type
	if i.Module != "" {
		label += fmt.Sprintf(".%s", i.Module)
	}
	if i.Resource != "" {
		label += fmt.Sprintf(".%s", i.Resource)
	}
	if i.Name != "" {
		label += fmt.Sprintf(".%s", i.Name)
	}

	return label
}

func readGraph(output string) *TerraformGraph {
	result := TerraformGraph{
		Items: make([]*TerraformGraphItem, 0),
	}

	var lines []string
	if strings.ContainsRune(output, '\r') {
		lines = strings.Split(output, "\r\n")
	} else {
		lines = strings.Split(output, "\n")
	}
	if len(lines) == 1 {
		return nil
	}

	initialIndex := 0
	if strings.HasPrefix(lines[0], "digraph {") {
		initialIndex += 1
	}
	if !strings.HasPrefix(lines[1], "compound =") {
		compound := strings.TrimSpace(lines[1])
		compound = strings.ReplaceAll(compound, "compound =", "")
		compound = strings.Trim(compound, "\"")
		if r, e := strconv.ParseBool(compound); e != nil {
			result.Compound = false
		} else {
			result.Compound = r
		}

		initialIndex += 1
	}
	if !strings.HasPrefix(lines[1], "newrank =") {
		newRank := strings.TrimSpace(lines[1])
		newRank = strings.ReplaceAll(newRank, "newrank =", "")
		newRank = strings.Trim(newRank, "\"")
		if r, e := strconv.ParseBool(newRank); e != nil {
			result.NewRank = false
		} else {
			result.NewRank = r
		}

		initialIndex += 1
	}
	if !strings.HasPrefix(lines[1], "subgraph =") {
		initialIndex += 1
	}
	if initialIndex >= len(lines) {
		return nil
	}

	for i := initialIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		// Getting resources first
		if !strings.Contains(line, "->") {
			item := TerraformGraphItem{
				DependsOn:  make([]string, 0),
				RequiredBy: make([]string, 0),
			}

			initialParts := strings.Split(line, "[")
			for _, part := range initialParts {
				parseRoot(part, &item)
				parseLabel(part, &item)
			}

			if item.Level != "" && item.Type != "" {
				result.Items = append(result.Items, &item)
			}
		}
	}

	for i := initialIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		// Getting dependencies in second to process
		if strings.Contains(line, "->") {
			dependencyParts := strings.Split(line, "->")
			if len(dependencyParts) == 2 {
				parsedItem := TerraformGraphItem{}
				parsedDependency := TerraformGraphItem{}
				parseRoot(dependencyParts[0], &parsedItem)
				parseRoot(dependencyParts[1], &parsedDependency)

				if parsedItem.Type != "" && parsedDependency.Type != "" {
					item := result.GetGraphItem(parsedItem.GetLabel())
					if item != nil {
						dItem := result.GetGraphItem(parsedDependency.GetLabel())
						if dItem != nil {
							item.DependsOn = append(item.DependsOn, dItem.Label)
							dItem.RequiredBy = append(dItem.RequiredBy, item.Label)
						}
					}
				}
			}
		}
	}

	for _, item := range result.Items {
		msg := fmt.Sprintf("level: %s", item.Level)
		if item.Module != "" {
			msg += fmt.Sprintf(", module: %s", item.Module)
		}
		if item.Resource != "" {
			msg += fmt.Sprintf(", resource: %s", item.Resource)
		}
		if item.Name != "" {
			msg += fmt.Sprintf(", name: %s", item.Name)
		}

		if len(item.DependsOn) > 0 {
			msg += "\n    dependsOn:\n"
			for i, d := range item.DependsOn {
				msg += fmt.Sprintf("      %s", d)
				if i != len(item.DependsOn)-1 {
					msg += "\n"
				}
			}
		}
		if len(item.RequiredBy) > 0 {
			msg += "\n    requiredBy:\n"
			for i, d := range item.RequiredBy {
				msg += fmt.Sprintf("      %s", d)
				if i != len(item.RequiredBy)-1 {
					msg += "\n"
				}
			}
		}

		notify.Debug(msg)
	}

	return &result
}

func parseRoot(value string, item *TerraformGraphItem) {
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "[")
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, "\"", "")
	if strings.HasPrefix(value, "root") {
		rootParts := strings.Split(value, " ")
		if len(rootParts) >= 2 {
			item.Level = strings.Trim(rootParts[0], "\"")
			resourceParts := strings.Split(rootParts[1], ".")
			if len(resourceParts) == 2 {
				item.Type = strings.Trim(resourceParts[0], "\"")
				item.Name = strings.Trim(resourceParts[1], "\"")
			}
			if len(resourceParts) == 3 {
				item.Type = strings.Trim(resourceParts[0], "\"")
				item.Resource = strings.Trim(resourceParts[1], "\"")
				item.Name = strings.Trim(resourceParts[2], "\"")
			}
			if len(resourceParts) == 4 {
				item.Type = strings.Trim(resourceParts[0], "\"")
				item.Module = strings.Trim(resourceParts[1], "\"")
				item.Resource = strings.Trim(resourceParts[2], "\"")
				item.Name = strings.Trim(resourceParts[3], "\"")
			} else if len(resourceParts) > 4 {
				item.Type = strings.Trim(resourceParts[0], "\"")
				item.Module = strings.Trim(resourceParts[1], "\"")
				item.Resource = strings.Trim(resourceParts[2], "\"")
				for i := 3; i < len(resourceParts); i++ {
					if len(item.Name) > 0 {
						item.Name += " "
					}

					item.Name += resourceParts[i]
				}
			}
		}
	}
}

func parseLabel(value string, item *TerraformGraphItem) {
	value = strings.Trim(value, "\"")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, "\"", "")
	if strings.HasPrefix(value, "label") {
		labelParts := strings.Split(value, ",")
		if len(labelParts) == 2 {
			label := strings.ReplaceAll(labelParts[0], "label = ", "")
			label = strings.Trim(label, "\"")
			label = strings.ReplaceAll(label, "\"", "")
			item.Label = label
		}
	}
}

func (s *TerraformGraph) GetGraphItem(label string) *TerraformGraphItem {
	for _, item := range s.Items {
		if strings.EqualFold(label, item.Label) {
			return item
		}
	}

	return nil
}
