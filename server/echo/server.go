package echo

import (
	"log"
	"net"
	"net/http"
	"path"
	"sync"

	"github.com/sevenNt/ares/server"
	"golang.org/x/net/websocket"
)

// ServerOption is used to set options for the server.
//type ServerOption func(*Server)

// HandlerFunc defines a function to server HTTP requests.
type HandlerFunc func(*Context) error

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Middleware is middleware interface for echo server.
type Middleware interface {
	Func() MiddlewareFunc
	//HookRoute(string, string)
	//Clone() (Middleware, bool)
}

// Handler is handler interface for echo server.
type Handler interface {
	Get(*Context) error
	Post(*Context) error
	Put(*Context) error
	Delete(*Context) error
}

// Server is the top-level framework echo server instance.
type Server struct {
	router          *Router
	listener        *Listener
	mws             []Middleware
	notFoundHandler HandlerFunc
	errHandler      HandlerFunc
	pool            *sync.Pool
	Func            func(*Server)
	addr            string
	name            string
	alias           []string
	//events             trace.EventLog
	hookersBeforeServe []func(*Server)
	grpcProxyWrapper   func(interface{}) HandlerFunc
	wsWrapper          func(h websocket.Handler) HandlerFunc

	running chan struct{}
}

// RouteInfo defines route information.
type RouteInfo struct {
	Method string
	Path   string
}

// NewServer constructs an instance of echo server.
func NewServer(lis net.Listener, opts ...server.Option) *Server {
	s := &Server{
		router: newRouter(),
		pool: &sync.Pool{
			New: func() interface{} {
				return newContext()
			},
		},
		//opts:             opts,
		mws:                make([]Middleware, 0),
		notFoundHandler:    notFoundHandler,
		errHandler:         errHandler,
		grpcProxyWrapper:   defaultGRPCProxyWrapper(),
		wsWrapper:          defaultWSWrapper(),
		listener:           wrapListener(lis),
		hookersBeforeServe: make([]func(*Server), 0),
		running:            make(chan struct{}, 1),
	}
	return s
}

// IsRunning checks which server is running.
func (s *Server) IsRunning() bool {
	return len(s.running) > 0
}

// GetListener returns server's listener.
func (s *Server) GetListener() *Listener {
	return s.listener
}

// Scheme returns server's scheme.
func (s *Server) Scheme() string {
	return "http"
}

// SetErrHandler sets server's errHandler.
func (s *Server) SetErrHandler(handler HandlerFunc) {
	s.errHandler = handler
}

// SetNotFoundHandler sets server's notFoundHandler.
func (s *Server) SetNotFoundHandler(handler HandlerFunc) {
	s.notFoundHandler = handler
}

// HookBeforeServe injects hooks executed before server run.
func (s *Server) HookBeforeServe(fn func(*Server)) {
	s.hookersBeforeServe = append(s.hookersBeforeServe, fn)
}

// Route routes echo server.
// Deprecated: user ares.Server() instead.
func (s *Server) Route(f func(*Server)) *Server {
	s.Func = f
	return s
}

// GetRouteInfos returns RouteInfos.
func (s *Server) GetRouteInfos() []RouteInfo {
	ret := make([]RouteInfo, 0)

	for _, r := range s.router.routes {
		ret = append(ret, RouteInfo{
			Method: r.method,
			Path:   r.path,
		})
	}

	return ret
}

// Addr return server address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Name return server name.
func (s *Server) Name() string {
	return s.name
}

// Alias returns server aliasname
func (s *Server) Alias() []string {
	return s.alias
}

// DumpRouteInfo dumps route infos.
func (s *Server) DumpRouteInfo() {
	for _, r := range s.router.routes {
		log.Printf("[ECHO] \x1b[34m%8s\x1b[0m %s", r.method, r.path)
	}
	log.Printf("[ECHO] \x1b[33m%8s\x1b[0m %s", "Listen On", s.Addr())
}

// Serve serves provided net listener.
func (s *Server) Serve() error {
	s.DumpRouteInfo()
	for _, hooker := range s.hookersBeforeServe {
		hooker(s)
	}
	s.running <- struct{}{}
	defer func() {
		<-s.running
	}()
	return http.Serve(s.listener, s)
}

// Stop stops server.
func (s *Server) Stop() {
	s.listener.Close()
}

// GracefulStop stops server graceful.
func (s *Server) GracefulStop() {
	s.listener.Close()
	s.listener.wg.Wait()
}

// ServeHTTP servers http server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.pool.Get().(*Context)
	defer s.pool.Put(c)
	c.reset(r, w, s)

	h := func(c *Context) error {
		var handler = s.notFoundHandler
		if node, pvalues := s.router.find(r.Method, r.URL.Path); node != nil {
			for i, pname := range node.pnames {
				c.params[pname] = pvalues[i]
			}
			if f, ok := node.fn.(HandlerFunc); ok {
				handler = f
				c.handler = handler
			}
			c.SetPatternPath(node.ppath)
		}

		return handler(c)
	}

	if err := h(c); err != nil {
		log.Printf("ServeHTTP error %s", err)
		s.errHandler(c)
	}

	c.Response().Flush()
	return
}

// Exit exits.
func (s *Server) Exit() error {
	return nil
}

// Group creates a new router group with prefix and optional group-level middleware.
func (s *Server) Group(prefix string, ms ...Middleware) *Group {
	return &Group{
		ms:     append(s.mws, ms...),
		server: s,
		path:   prefix,
	}
}

// Handle registers Restful Operators.
func (s *Server) Handle(path string, h Handler, ms ...Middleware) *Server {
	s.add(GET, path, h.Get, ms...)
	s.add(PUT, path, h.Put, ms...)
	s.add(POST, path, h.Post, ms...)
	s.add(DELETE, path, h.Delete, ms...)
	return s
}

// Use adds middleware to the chain which is run after router.
func (s *Server) Use(ms ...Middleware) *Server {
	s.mws = append(s.mws, ms...)
	return s
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (s *Server) GET(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(GET, path, handler, append(s.mws, ms...)...)
	return s
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (s *Server) POST(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(POST, path, handler, append(s.mws, ms...)...)
	return s
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (s *Server) PUT(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(PUT, path, handler, append(s.mws, ms...)...)
	return s
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (s *Server) HEAD(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(HEAD, path, handler, append(s.mws, ms...)...)
	return s
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (s *Server) OPTIONS(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(OPTIONS, path, handler, append(s.mws, ms...)...)
	return s
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (s *Server) PATCH(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(PATCH, path, handler, append(s.mws, ms...)...)
	return s
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (s *Server) DELETE(path string, handler HandlerFunc, ms ...Middleware) *Server {
	s.add(DELETE, path, handler, append(s.mws, ms...)...)
	return s
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (s *Server) Static(prefix, root string) *Server {
	s.add(GET, prefix+"/:path", func(c *Context) error {
		return c.File(path.Join(root, c.Param("path")))
	}, s.mws...)
	return s
}

// WS registers a new websocket route for a path with matching handler in the router
// with optional route-level middleware.
func (s *Server) WS(path string, handler websocket.Handler, ms ...Middleware) *Server {
	//s.add2(GET, path, handler, append(s.mws, ms...)...)
	s.add(GET, path, s.wsWrapper(handler), ms...)
	return s
}

// GRPCProxy will appoint a http proxy for gRPC callers with POST method, ex. s.GRPC("sayhello", new(Greeter).SayHello).
func (s *Server) GRPCProxy(path string, handler interface{}, ms ...Middleware) *Server {
	s.add(POST, path, s.grpcProxyWrapper(handler), ms...)
	return s
}

// GRPCProxyGet will appoint a http proxy for gRPC callers with GET method, ex. s.GRPC("sayhello", new(Greeter).SayHello).
func (s *Server) GRPCProxyGet(path string, handler interface{}, ms ...Middleware) *Server {
	s.add(GET, path, s.grpcProxyWrapper(handler), ms...)
	return s
}

// GRPCProxyPost do same thing with GRPCProxy.
func (s *Server) GRPCProxyPost(path string, handler interface{}, ms ...Middleware) *Server {
	s.add(POST, path, s.grpcProxyWrapper(handler), ms...)
	return s
}

func (s *Server) add(method, path string, handler HandlerFunc, ms ...Middleware) {
	h := handler

	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i].Func()(h)
	}
	s.router.add(method, path, h)
}
