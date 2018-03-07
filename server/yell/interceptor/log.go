package interceptor

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCContextHeader returns grpc's header.
func GRPCContextHeader(ctx context.Context, key string) (value string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return strings.Join(md[strings.ToLower(key)], " ")
	}

	return
}

// Log logs interceptor.
func Log(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vs := range md {
			fmt.Println(k, " : ", strings.Join(vs, " "))
		}
	}

	return handler(ctx, req)
}
