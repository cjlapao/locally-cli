package interfaces

type EnvironmentVaultItem interface {
	GetKey() string
	GetValue() string
	IsEncrypted() bool
	IsSecret() bool
}
