package models

// ServiceDefinition defines a service and its modules
type ServiceDefinition struct {
	Name        string                       `json:"name" yaml:"name"`
	Description string                       `json:"description" yaml:"description"`
	Modules     map[string]*ModuleDefinition `json:"modules" yaml:"modules"`
}

// ModuleDefinition defines a module and its allowed actions
type ModuleDefinition struct {
	System      string        `json:"-" yaml:"-"`
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Actions     []AccessLevel `json:"actions" yaml:"actions"`
}

// DefaultAccessLevels defines the default access levels for each security level
type DefaultAccessLevels struct {
	SuperUser AccessLevel `json:"superuser" yaml:"superuser"`
	Admin     AccessLevel `json:"admin" yaml:"admin"`
	Manager   AccessLevel `json:"manager" yaml:"manager"`
	User      AccessLevel `json:"user" yaml:"user"`
	Guest     AccessLevel `json:"guest" yaml:"guest"`
	None      AccessLevel `json:"none" yaml:"none"`
}
