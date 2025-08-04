package models

type Service struct {
	Name        string
	Description string
	Modules     []Module
}

type Module struct {
	Name        string
	Description string
	Actions     []AccessLevel
}
