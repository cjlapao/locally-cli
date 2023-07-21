package interfaces

type LneService interface {
	GetName() string
	GetDependencies() []string
	GetSource() string
	AddDependency(value string)
	AddRequiredBy(value string)
	BuildDependency() error
	SaveFragment() error
}
