package configuration

import (
	"bytes"
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v3"
)

type PipelineTaskType uint

const (
	UnknownTask PipelineTaskType = iota
	InfrastructureTask
	SqlTask
	BashTask
	CurlTask
	EFMigrationTask
	DotnetTask
	DockerTask
	ProxyTask
	GitTask
	KeyvaultSyncTask
	EmsTask
	WhatsNewTask
	NpmTask
  WebClientManifestTask
)

var toPipelineTaskTypeString = map[PipelineTaskType]string{
	BashTask:              "bash",
	CurlTask:              "curl",
	EFMigrationTask:       "migrations",
	InfrastructureTask:    "infrastructure",
	SqlTask:               "sql",
	DockerTask:            "docker",
	ProxyTask:             "proxy",
	GitTask:               "git",
	KeyvaultSyncTask:      "keyvault",
	DotnetTask:            "dotnet",
	EmsTask:               "ems",
	WhatsNewTask:          "whatsnew",
	WebClientManifestTask: "webclientmanifest",
	UnknownTask:           "unknown",
	NpmTask:               "npm",
}

var toPipelineTaskType = map[string]PipelineTaskType{
	"bash":              BashTask,
	"curl":              CurlTask,
	"migrations":        EFMigrationTask,
	"infrastructure":    InfrastructureTask,
	"sql":               SqlTask,
	"docker":            DockerTask,
	"proxy":             ProxyTask,
	"git":               GitTask,
	"keyvault":          KeyvaultSyncTask,
	"dotnet":            DotnetTask,
	"ems":               EmsTask,
	"whatsnew":          WhatsNewTask,
	"webclientmanifest": WebClientManifestTask,
	"unknown":           UnknownTask,
  "npm":               NpmTask,
}

func (t PipelineTaskType) String() string {
	return toPipelineTaskTypeString[t]
}

func (t PipelineTaskType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toPipelineTaskTypeString[t])
	buffer.WriteString(`"`)

	return buffer.Bytes(), nil
}

func (t *PipelineTaskType) UnmarshalJSON(b []byte) error {
	var key string
	err := json.Unmarshal(b, &key)
	if err != nil {
		return err
	}

	*t = toPipelineTaskType[key]
	return nil
}

func (t PipelineTaskType) MarshalYAML() (interface{}, error) {
	notify.Debug("%v", t)
	return toPipelineTaskTypeString[t], nil
}

func (t *PipelineTaskType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.Kind(yaml.LiteralStyle):
		*t = toPipelineTaskType[value.Value]
	default:
		return errors.New("invalid format for the current enum type")
	}
	return nil
}
