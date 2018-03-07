package chiwoo

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type (
	ClientInterceptor interface {
		UnaryClientIntercept() grpc.UnaryClientInterceptor
		StreamClientIntercept() grpc.StreamClientInterceptor
	}
)

// Chiwoo 一个grpc客户端的wrapper
type Chiwoo struct {
	target             string
	dialOptions        []grpc.DialOption
	dialTimeout        time.Duration
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
}

// New 返回一个grpc client wrappter Chiwoo
func New(target string, opts ...grpc.DialOption) *Chiwoo {
	return &Chiwoo{
		dialOptions:        make([]grpc.DialOption, 0),
		dialTimeout:        time.Second * 2,
		unaryInterceptors:  make([]grpc.UnaryClientInterceptor, 0),
		streamInterceptors: make([]grpc.StreamClientInterceptor, 0),
	}
}

func (c *Chiwoo) SetTarget(target string) *Chiwoo {
	c.target = target
	return c
}

func (c *Chiwoo) SetDialOptions(opts ...grpc.DialOption) *Chiwoo {
	c.dialOptions = append(c.dialOptions, opts...)
	return c
}

func (c *Chiwoo) SetUnaryClientInterceptor(intes ...ClientInterceptor) *Chiwoo {
	for _, in := range intes {
		c.unaryInterceptors = append(c.unaryInterceptors, in.UnaryClientIntercept())
	}
	return c
}

func (c *Chiwoo) SetStreamClientInterceptor(intes ...ClientInterceptor) *Chiwoo {
	for _, in := range intes {
		c.streamInterceptors = append(c.streamInterceptors, in.StreamClientIntercept())
	}
	return c
}

func (c *Chiwoo) SetETCDBalancerEndpoints(endpoints ...string) *Chiwoo {
	c.dialOptions = append(c.dialOptions, grpc.WithBalancer(NewETCDBalancer(endpoints)))
	return c
}

func (c *Chiwoo) SetBalancer(b grpc.Balancer) *Chiwoo {
	c.dialOptions = append(c.dialOptions, grpc.WithBalancer(b))
	return c
}

func (c *Chiwoo) SetDialTimeout(timeout time.Duration) *Chiwoo {
	c.dialTimeout = timeout
	return c
}

func (c *Chiwoo) MustDial() *grpc.ClientConn {
	conn, err := c.Dial()
	if err != nil {
		panic(err)
	}

	return conn
}

func (c *Chiwoo) Dial() (cc *grpc.ClientConn, err error) {
	var ctx = context.Background()
	if c.dialTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
	}

	if len(c.unaryInterceptors) > 0 {
		c.dialOptions = append(c.dialOptions, grpc.WithUnaryInterceptor(UnaryClientInterceptorChain(c.unaryInterceptors...)))
	}

	if len(c.streamInterceptors) > 0 {
		c.dialOptions = append(c.dialOptions, grpc.WithStreamInterceptor(StreamClientInterceptorChain(c.streamInterceptors...)))
	}

	return grpc.DialContext(ctx, c.target, c.dialOptions...)
}
