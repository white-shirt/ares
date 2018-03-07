package echo

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/codegangsta/inject"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

/*
 * 对Restful接口来说，你可以实现Handler接口，定义自己的base handler
 * 通过在自定义的base handler中添加钩子函数，实现更灵活的功能
 */

const (
	codeOK                   = 0
	codeMS                   = 10000
	codeMSInvalidParam       = 10001
	codeMSInvoke             = 10002
	codeMSInvokeLen          = 10003
	codeMSSecondItemNotError = 10004
	codeMSResErr             = 10005
)

var (
	errBadRequest         = status.Errorf(codes.InvalidArgument, createStatusErr(codeMSInvalidParam, "bad request"))
	errMicroDefault       = status.Errorf(codes.Internal, createStatusErr(codeMS, "micro default"))
	errMicroInvoke        = status.Errorf(codes.Internal, createStatusErr(codeMSInvoke, "invoke failed"))
	errMicroInvokeLen     = status.Errorf(codes.Internal, createStatusErr(codeMSInvokeLen, "invoke result not 2 item"))
	errMicroInvokeInvalid = status.Errorf(codes.Internal, createStatusErr(codeMSSecondItemNotError, "second invoke res not a error"))
	errMicroResInvalid    = status.Errorf(codes.Internal, createStatusErr(codeMSResErr, "response is not valid"))
)

// BaseHandler defines base handler for restful api.
type BaseHandler struct{}

// Get Get.
func (h *BaseHandler) Get(c *Context) error {
	return notFoundHandler(c)
}

// Post Post.
func (h *BaseHandler) Post(c *Context) error {
	return notFoundHandler(c)
}

// Put Put.
func (h *BaseHandler) Put(c *Context) error {
	return notFoundHandler(c)
}

// Delete Delete.
func (h *BaseHandler) Delete(c *Context) error {
	return notFoundHandler(c)
}

func notFoundHandler(c *Context) error {
	return c.String(StatusNotFound, http.StatusNotFound)
}

func errHandler(c *Context) error {
	return c.String(StatusInternalServerError, http.StatusInternalServerError)
}

func defaultWSWrapper() func(h websocket.Handler) HandlerFunc {
	return func(h websocket.Handler) HandlerFunc {
		return func(c *Context) error {
			h.ServeHTTP(c.Response().ResponseWriter, c.Request().Request)
			return nil
		}
	}
}

func defaultGRPCProxyWrapper() func(interface{}) HandlerFunc {
	return func(h interface{}) HandlerFunc {
		t := reflect.TypeOf(h)
		if t.Kind() != reflect.Func {
			panic("reflect error: handler must be func")
		}
		return func(c *Context) error {
			var req = reflect.New(t.In(1).Elem()).Interface()
			if err := c.Bind(req); err != nil {
				return c.ProtoError(StatusBadRequest, errBadRequest)
			}
			ctx := metadata.NewOutgoingContext(c.Context, metadata.MD(c.Request().Header()))
			var inj = inject.New()
			inj.Map(ctx)
			inj.Map(req)
			vs, err := inj.Invoke(h)
			if err != nil {
				// maybe some log
				return c.ProtoError(StatusInternalServerError, errMicroInvoke)
			}

			if len(vs) != 2 {
				return c.ProtoError(StatusInternalServerError, errMicroInvokeLen)
			}
			repV, errV := vs[0], vs[1]
			if !errV.IsNil() || repV.IsNil() {
				if e, ok := errV.Interface().(error); ok {
					// error logic
					return c.ProtoError(StatusOK, e)
				}
				return c.ProtoError(StatusInternalServerError, errMicroInvokeInvalid)
			}

			if !repV.IsValid() {
				// maybe some log
				return c.ProtoError(StatusInternalServerError, errMicroResInvalid)
			}

			return c.ProtoJSON(StatusOK, repV.Interface())
		}
	}
}

func createStatusErr(code uint32, msg string) string {
	return fmt.Sprintf("%d:%s", code, msg)
}
