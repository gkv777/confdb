package confdb

// Service ...
type Service interface {
	HealthCheck() bool
	SayHello(name string) string
}

// Srv ...
type service struct {}

// NewService ...
func NewService() Service {
	return service{}
}

// HealthCheck ...
func (service) HealthCheck() bool {
	return true
}

// SayHello ...
func (service) SayHello(name string) string {
	return "Hello " + name
}

