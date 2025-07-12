package interfaces

type RemoteContextProvider interface {
	Id() string
	Name() string
	TestConnection() error
}
