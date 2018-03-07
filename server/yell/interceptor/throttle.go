package interceptor

import (
	"time"

	"github.com/juju/ratelimit"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Throttle ...
type Throttle struct {
	FillInterval    time.Duration
	Capacity        int64
	Rate            float64
	WaitMaxDuration time.Duration
}

// Func ...
func (t Throttle) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	rt := ratelimit.NewBucket(t.FillInterval, t.Capacity)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, wait := rt.TakeMaxDuration(1, t.WaitMaxDuration); !wait {
			return nil, nil
		}
		return handler(ctx, req)
	}
}
