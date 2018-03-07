package echo

import "google.golang.org/grpc"
import "google.golang.org/grpc/codes"

// HTTPError wraps handler error.
type HTTPError struct {
	Code    int
	Message string
}

// NewHTTPError constructs a new HTTPError instance.
func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{Code: code, Message: StatusText(code)}
	if len(msg) > 0 {
		he.Message = msg[0]
	}

	return he
}

// Error return error message.
func (e HTTPError) Error() string {
	return e.Message
}

// ErrNotFound defines StatusNotFound error.
var ErrNotFound = HTTPError{
	Code:    StatusNotFound,
	Message: "not found",
}

var (
	ErrGRPCResponseValid = grpc.Errorf(codes.Internal, "response valid")
	ErrGRPCInvokeLen     = grpc.Errorf(codes.Internal, "invoke request without len 2 res")
)
