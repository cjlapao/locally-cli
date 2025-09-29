package types

type EnvironmentType string

const (
	EnvironmentTypeGlobal  EnvironmentType = "global"
	EnvironmentTypeTenant  EnvironmentType = "tenant"
	EnvironmentTypeProject EnvironmentType = "project"
)

type EnvironmentVaultItemType string

const (
	EnvironmentVaultItemTypeString  EnvironmentVaultItemType = "string"
	EnvironmentVaultItemTypeNumber  EnvironmentVaultItemType = "number"
	EnvironmentVaultItemTypeBoolean EnvironmentVaultItemType = "boolean"
	EnvironmentVaultItemTypeArray   EnvironmentVaultItemType = "array"
	EnvironmentVaultItemTypeJSON    EnvironmentVaultItemType = "json"
)
