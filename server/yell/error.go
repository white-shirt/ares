package yell

import "google.golang.org/grpc/codes"

// A Code is an unsigned 32-bit error code as defined in the gRPC spec.
type Code = codes.Code

const (
	CodeTooManyRequest Code = 100
	CodeCircuitBreak   Code = 101
)

var strToCode = map[string]Code{
	`"TOO_MANY_REQUEST"`: CodeTooManyRequest,
}
