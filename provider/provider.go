package provider

type ServerEndpoint struct {
	ID      string
	Host    string
	Port    int
	CPUName string
	GPUName string
}

type LLMProvider interface {
	GetEndpoints() ([]ServerEndpoint, error)
	AutoScaling(replica int) error
	GetModel() string
}
