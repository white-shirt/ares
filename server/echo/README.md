# ECHO

ECHO 是一个基于[echo](https://github.com/labstack/echo)的高性能web服务框架.

#### 主要包含以下改进:
* 支持ares集成
* 增加Middleware Interface, 支持MiddelwareFunc克隆
* binder支持[govalidator](https://github.com/asaskevich/govalidator)
* 支持lazy render, 可通过middleware对响应结果整形
* 支持protobuf render
* Graceful Shutdown
* 支持Restful API
* 支持Grpc Handler转接

> echo可以单独使用，但建议与ares一同使用，ares提供了一些有用的服务治理组件.

## Startup

### Router

> GET POST PUT PATCH DELETE OPTIONS etc.

```go
func route(s *echo.Server) {
    // 直接注册路由
    s.GET("/user/:id/@info", user.Info, mw)
    s.POST("/user", user.Post, mw)
    s.PUT("/user/:id", user.Update, mw)
    s.PATCH("/user/:id/@addr", user.UpdateAddr, mw)
    s.DELETE("/user/:id", user.Delete, mw)

    // 或者通过路由组
    group := s.Group("/user")
    {
        group.GET("/:id/@info", user.Info, mw)
        group.POST("/", user.Post, mw)
        group.PUT("/:id", user.Update, mw)
        group.PATCH("/:id/@addr", user.UpdateAddr, mw)
        group.DELETE("/:id", user.Delete, mw)
    }

    // 转接到grpc handler
    // 注意: 只转接到handler，grpc拦截器不能生效，需自行注入http middleware
    s.GRPCProxy("/greeter/sayhello", greeter.SayHello, mw)
    // 或者
    {
        group.GRPCProxy("/greeter/sayhello", greeter.SayHello, mw)
    }
}
```

### Binder

> JSON XML Form Query FormPost FormMultipart Protobuf
Binder集成[govalidator](https://github.com/asaskevich/govalidator)
validate tag可参考govalidator文档

Binder自动根据Content-Type选择binder，对于请求header中没有指定Content-Type的，默认使用FormBinder; 详细绑定对应关系如下:

``` go
	switch contentType {
	case MIMEApplicationJSON, MIMEApplicationJSONCharsetUTF8:
		return JSONBinder
	case MIMEApplicationXML, MIMEApplicationXMLCharsetUTF8:
		return XMLBinder
	case MIMEApplicationProtobuf:
		return ProtoBufBinder
	case MIMEMultipartForm:
		return FormMultipartBinder
	case MIMEApplicationForm:
		return FormBinder
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		return FormBinder
	}
```

```go
type UserInfo struct {
    Name    string  `valid:"-"`
    Age     int     `valid:""`
    Email   string  `valid:"email"`
}
func (u User) Info(c *echo.Context) error {
    userInfo := UserInfo{}
    if err := c.Bind(&userInfo); err != nil {
        // do something
    }
}
```

### Render

> JSON XML Blob Protobuf Msgpack HTML

