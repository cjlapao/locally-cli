package interfaces

type Operation interface {
	GetName() string
	Run(arguments ...string)
}
