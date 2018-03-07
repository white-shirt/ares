package echo

import (
	"strings"

	"golang.org/x/net/websocket"
)

// Group is a set of sub-routes for a specified route. It can be used for inner
// routes that share a common middleware or functionality that should be separate
// from the parent echo instance while still inheriting from it.
type Group struct {
	server *Server
	ms     []Middleware
	path   string
}

// Use implements `Echo#Use()` for sub-routes within the Group.
func (g *Group) Use(ms ...Middleware) {
	g.ms = append(g.ms, ms...)
}

// GET implements `Echo#GET()` for sub-routes within the Group.
func (g *Group) GET(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(GET, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// PUT implements `Echo#PUT()` for sub-routes within the Group.
func (g *Group) PUT(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(PUT, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// POST implements `Echo#POST()` for sub-routes within the Group.
func (g *Group) POST(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(POST, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
func (g *Group) DELETE(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(DELETE, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
func (g *Group) PATCH(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(PATCH, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
func (g *Group) OPTIONS(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(OPTIONS, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
func (g *Group) HEAD(path string, handler HandlerFunc, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(HEAD, g.path+path, handler, append(g.ms, ms...)...)
	return g
}

// WS implements `Echo#WS()` for sub-routes within the Group.
func (g *Group) WS(path string, handler websocket.Handler, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(GET, g.path+path, g.server.wsWrapper(handler), append(g.ms, ms...)...)
	return g
}

func (g *Group) GRPCProxy(path string, handler interface{}, ms ...Middleware) *Group {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.server.add(POST, g.path+path, g.server.grpcProxyWrapper(handler), append(g.ms, ms...)...)
	return g
}
