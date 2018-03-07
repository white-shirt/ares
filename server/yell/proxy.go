package yell

type Endpoint struct {
	Req interface{}
	Rep interface{}
	Fun interface{}
}

type Proxy interface {
	Proxy() map[string]Endpoint
}
