package registry

// Service description
type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Nodes     []*Node           `json:"nodes"`
	Endpoints []*Endpoint       `json:"endpoints"`
}

// Node service node description
type Node struct {
	ID       string            `json:"id"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
}

// Endpoint ...
type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

// Value ...
type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}
