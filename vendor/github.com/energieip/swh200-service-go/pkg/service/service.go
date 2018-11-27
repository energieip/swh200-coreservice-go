package service

// Service definition
type Service interface {
	Initialize(confFile string) error
	Stop()
	Run() error
}
