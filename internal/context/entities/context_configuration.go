package entities

import "github.com/cjlapao/locally-cli/internal/entities"

type ContextConfiguration struct {
	Domain               string                         `json:"domain,omitempty" yaml:"domain,omitempty"`
	Subdomain            string                         `json:"subdomain,omitempty" yaml:"subdomain,omitempty"`
	SchemaVersion        string                         `json:"schemaVersion,omitempty" yaml:"schemaVersion,omitempty"`
	Location             *ContextLocation               `json:"location,omitempty" yaml:"location,omitempty"`
	LocationType         string                         `json:"locationType,omitempty" yaml:"locationType,omitempty"`
	LocationUsername     string                         `json:"locationUsername,omitempty" yaml:"locationUsername,omitempty"`
	LocationPassword     string                         `json:"locationPassword,omitempty" yaml:"locationPassword,omitempty"`
	ConfigFolder         string                         `json:"folder,omitempty" yaml:"folder,omitempty"`
	RootURI              string                         `json:"rootUri,omitempty" yaml:"rootUri,omitempty"`
	OutputPath           string                         `json:"outputPath,omitempty" yaml:"outputPath,omitempty"`
	LocallyConfigService *entities.LocallyConfigService `json:"locallyConfigService,omitempty" yaml:"locallyConfigService,omitempty"`
}

type ContextLocation struct {
	Path     string `json:"path,omitempty" yaml:"path,omitempty"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}
