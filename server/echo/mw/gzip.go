package mw

import (
	"compress/gzip"
	"strings"

	"github.com/sevenNt/ares/server/echo"
)

// DefaultGZip default gzip.
var DefaultGZip = GZip{
	Level: 1,
}

// GZip gzip middleware.
type GZip struct {
	Base
	Level int
}

// Func implements Middleware interface.
func (m GZip) Func() echo.MiddlewareFunc {
	if m.Level == 0 {
		m.Level = DefaultGZip.Level
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if strings.Contains(c.Request().Header().Get(echo.HeaderAcceptEncoding), "gzip") {
				c.Response().Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
				c.Response().Header().Add(echo.HeaderContentEncoding, "gzip")

				gw, err := gzip.NewWriterLevel(c.Response().ResponseWriter, m.Level)
				if err != nil {
					return err
				}
				c.Response().SetWriter(&gzipResponseWriter{
					w: gw,
				})
			}
			return next(c)
		}
	}
}

type gzipResponseWriter struct {
	w *gzip.Writer
}

func (w *gzipResponseWriter) Write(bs []byte) (n int, e error) {
	if num, err := w.w.Write(bs); err != nil {
		return num, err
	}

	if err := w.w.Flush(); err != nil {
		return -1, err
	}

	return
}
