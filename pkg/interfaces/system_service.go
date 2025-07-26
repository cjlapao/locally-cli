package interfaces

type SystemService[T any] interface {
	GetInstance() *SystemService[T]
	Initialize() *SystemService[T]
	GetName() string
}
