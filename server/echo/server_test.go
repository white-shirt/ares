package echo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/sevenNt/ares/server/echo"
)

func TestRoute(t *testing.T) {
	server := echo.New()
	server.GET("/ping", func(c *echo.Context) error {
		return c.JSON(200, nil)
	})

	Convey("GET /ping", t, func() {
		req := httptest.NewRequest("GET", "/ping", nil)
		res := httptest.NewRecord()

		server.ServeHTTP(res, req)
		So(res.Result().StatusCode(), ShouldEqual, 200)
	})
}

func TestNewServer(t *testing.T) {
	//server := echo.New()
	//server.GET("/ping", func(c *echo.Context) error {
	//return c.JSON(200, nil)
	//})
	//server.Run()
}
