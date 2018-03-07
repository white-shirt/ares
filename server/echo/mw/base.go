package mw

import (
	"github.com/sevenNt/ares/server/echo"
)

type Base struct{}

func (b Base) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			return next(c)
		}
	}
}

// HookRoutePath 在构建HandlerFunc闭包的时候执行，用于根据URL Path初始化中间件
// 一般cache metric tracing等关心url path的中间件可重载该方法
func (b Base) HookRoute(method, path string) {
	return
}

// 默认不克隆
func (b Base) Clone() (echo.Middleware, bool) {
	return nil, false
}

type MW struct {
	Base
	h func(*echo.Context, echo.HandlerFunc) error
}

func New(h func(*echo.Context, echo.HandlerFunc) error) *MW {
	return &MW{h: h}
}

func (m MW) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			return m.h(c, next)
		}
	}
}
