package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sevenNt/ares/codec"
	rstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

// Context represents the context of the current HTTP request. It holds request and
// response objects, path, path parameters, data and registered handler.
type Context struct {
	context.Context
	id       string
	request  *Request
	response *Response
	handler  HandlerFunc
	params   map[string]string
	render   Render
	reqPool  sync.Pool
	repPool  sync.Pool
	path     string // request path: /v1/user/123
	ppath    string // pattern path: /v1/user/:id
}

func newContext() *Context {
	return &Context{}
}

func (c *Context) reset(req *http.Request, res http.ResponseWriter, s *Server) {
	c.Context = context.Background()
	c.request = newRequest(req)
	c.response = newResponse(res)
	c.params = make(map[string]string)
	// TODO server\writer\render\handler\id
}

// ID returns context ID.
func (c *Context) ID() string {
	return c.id
}

// Bind binds the request body into provided type `i`. The default binder
// does it based on Content-Type header.
func (c *Context) Bind(i interface{}) error {
	b := defaultBinder(c.request.Method, c.request.Header().Get(HeaderContentType))
	return c.BindWith(i, b)
}

// BindWith binds the request body into provided type `i`. The default binder
// does it based on Content-Type header with provided binder.
func (c *Context) BindWith(obj interface{}, binder Binder) error {
	return binder.Bind(c.request, obj)
}

// ClientIP returns client IP.
func (c *Context) ClientIP() string {
	for _, ip := range c.clientIPs() {
		// not ipv4, ipv6
		if !net.ParseIP(strings.TrimSpace(ip)).IsGlobalUnicast() {
			continue
		}

		return ip
	}

	return "0.0.0.0"
}

func (c *Context) clientIPs() []string {
	if ips := c.request.Header().Get("Cdn-Src-Ip"); ips != "" {
		return strings.Split(ips, ",")
	}
	if ips := c.request.Header().Get(HeaderXForwardedFor); ips != "" {
		return strings.Split(ips, ",")
	}
	if ips := c.request.Header().Get(HeaderXRealIP); ips != "" {
		return strings.Split(ips, ",")
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.request.RemoteAddr)); err == nil {
		return []string{ip}
	}
	return []string{}
}

// ContentType returns request header content type.
func (c *Context) ContentType() string {
	return c.request.Header().Get(HeaderContentType)
}

// SetContext sets provided context.
func (c *Context) SetContext(ctx context.Context) {
	c.Context = ctx
}

// Deadline returns context deadline.
func (c *Context) Deadline() (time.Time, bool) {
	return c.Context.Deadline()
}

// Done returns context done channel.
func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

// Err returns context error.
func (c *Context) Err() error {
	return c.Context.Err()
}

// Value returns context value by provided key.
func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

// Request returns context request.
func (c *Context) Request() *Request {
	return c.request
}

// Response returns context response.
func (c *Context) Response() *Response {
	return c.response
}

// Path returns context path.
func (c *Context) Path() string {
	return c.path
}

// SetPath sets context path.
func (c *Context) SetPath(p string) {
	c.path = p
}

// PatternPath returns context path.
func (c *Context) PatternPath() string {
	return c.ppath
}

// SetPatternPath sets context path.
func (c *Context) SetPatternPath(p string) {
	c.ppath = p
}

// SetHandler sets context handler.
func (c *Context) SetHandler(h HandlerFunc) {
	c.handler = h
}

// Param returns path parameter by name.
func (c *Context) Param(name string) (value string) {
	return c.params[name]
}

// ParamInt returns path int parameter by name.
func (c *Context) ParamInt(name string, expect int) (value int) {
	if res, err := strconv.ParseInt(c.Param(name), 10, 64); err == nil {
		return int(res)
	}
	return expect
}

// Query returns the query param for the provided name.
func (c *Context) Query(name string) string {
	return c.request.Query(name)
}

// QueryS returns the URL query strings.
func (c *Context) QueryS(name string) []string {
	return c.request.QueryS(name)
}

// QueryAll returns all queries.
func (c *Context) QueryAll() url.Values {
	return c.request.URL.QueryAll()
}

// QueryString returns the URL query string.
func (c *Context) QueryString(name string, expect string) string {
	if res := c.Query(name); res != "" {
		return res
	}

	return expect
}

// QueryUint returns uint query by name.
func (c *Context) QueryUint(name string, expect uint64) uint64 {
	if res, err := strconv.ParseUint(c.Query(name), 10, 64); err == nil {
		return res
	}

	return expect
}

// QueryInt returns int query by name.
func (c *Context) QueryInt(name string, expect int64) int64 {
	if res, err := strconv.ParseInt(c.Query(name), 10, 64); err == nil {
		return res
	}

	return expect
}

// QueryBool returns bool query by name.
func (c *Context) QueryBool(name string, expect bool) bool {
	if res, err := strconv.ParseBool(c.Query(name)); err == nil {
		return res
	}

	return expect
}

// QueryFloat64 returns float64 query by name.
func (c *Context) QueryFloat64(name string, expect float64) float64 {
	if res, err := strconv.ParseFloat(c.Query(name), 64); err == nil {
		return res
	}

	return expect
}

// FormValue returns the form field value for the provided name.
func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

// FormFile returns the multipart form file for the provided name.
func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.request.FormFile(name)
}

// FormParams returns the form parameters as `url.Values`.
func (c *Context) FormParams() (url.Values, error) {
	const defaultMemory = 32 << 20
	if strings.HasPrefix(c.request.Header().Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}

	return c.request.Form, nil
}

// MultipartForm returns the multipart form.
func (c *Context) MultipartForm() *multipart.Form {
	return c.request.MultipartForm
}

// Cookie returns the named cookie provided in the request.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *Context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

// Set saves data in the context.
// TODO should not use basic type string as key in context.WithValue
func (c *Context) Set(key string, val interface{}) {
	c.Context = context.WithValue(c.Context, key, val)
}

// Get retrieves data from the context.
func (c *Context) Get(key string) interface{} {
	return c.Context.Value(key)
}

// File sends a response with the content of the file.
func (c *Context) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	return c.ServeContent(f, fi.Name(), fi.ModTime())
}

// Attachment sends a response as attachment, prompting client to save the file.
func (c *Context) Attachment(r io.ReadSeeker, name string) (err error) {
	c.response.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	c.response.Header().Set(HeaderContentDisposition, "attachment; filename="+name)
	c.response.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.response, r)
	return
}

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(HeaderLocation, url)
	c.response.WriteHeader(code)
	return nil
}

// ServeContent replies to the request using the content in the provided ReadSeeker.
func (c *Context) ServeContent(content io.ReadSeeker, name string, modtime time.Time) error {
	req := c.request
	res := c.response

	if t, err := time.Parse(http.TimeFormat, req.Header().Get(HeaderIfModifiedSince)); err == nil && modtime.Before(t.Add(1*time.Second)) {
		res.Header().Del(HeaderContentType)
		res.Header().Del(HeaderContentLength)
		return c.NoContent(http.StatusNotModified)
	}

	res.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	res.Header().Set(HeaderLastModified, modtime.UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	_, err := io.Copy(res, content)
	return err
}

// ContentTypeByExtension returns the MIME type associated with the file based on
// its extension. It returns `application/octet-stream` incase MIME type is not
// found.
func ContentTypeByExtension(name string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
		t = MIMEOctetStream
	}
	return
}

// Render renders a response with provided render.
func (c *Context) Render(code int, r Render) (err error) {
	return r.Render(code, c)
}

// Status returns response status code.
func (c *Context) Status(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// String sends a string response with status code.
func (c *Context) String(code int, data interface{}) (err error) {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	}
	//return c.Render(code, stringRender{Data: data})
	c.response.SetStatus(code)
	c.response.SetData(data)
	c.response.SetEncoder(codec.NewStringCodec())
	return
}

// Stringf sends a string response with status code in specific format.
func (c *Context) Stringf(code int, format string, data ...interface{}) (err error) {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	}

	//return c.Render(code, stringRender{Data: data})
	c.response.SetStatus(code)
	if len(data) == 1 {
		c.response.SetData(data[0])
	} else {
		c.response.SetData(data)
	}
	c.response.SetEncoder(codec.NewStringCodec(codec.Format(format)))

	return
}

// JSON sends a JSON response with status code.
func (c *Context) JSON(code int, i interface{}) (err error) {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	}

	c.response.SetStatus(code)
	c.response.SetData(i)
	pretty := c.QueryBool("_pretty", false)
	c.response.SetEncoder(codec.NewJSONCodec(codec.Indent(pretty)))
	return nil
}

// JSONBlob sends a JSON blob response with status code.
func (c *Context) JSONBlob(code int, bs []byte) (err error) {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	}

	c.response.WriteHeader(code)
	_, err = c.response.Write(bs)

	return
}

// MsgPack pack msg with provided status code and data.
func (c *Context) MsgPack(code int, i interface{}) error {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMEApplicationMsgpack)
	}

	return c.Render(code, msgpackRender{Data: i})
}

// Protobuf sends a Protobuf response with status code.
func (c *Context) Protobuf(code int, i interface{}) error {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMEApplicationProtobuf)
	}

	return c.Render(code, protobufRender{Data: i})
}

// ProtoError sends a Protobuf JSON response with status code and error.
func (c *Context) ProtoError(code int, e error) error {
	s, ok := status.FromError(e)
	c.response.Header().Set(HeaderHRPCErr, "true")
	if ok {
		if de, ok := statusFromString(s.Message()); ok {
			return c.ProtoJSON(code, de.Proto())
		}
	}
	return c.ProtoJSON(code, e)
}

// ProtoJSON sends a Protobuf JSON response with status code and data.
func (c *Context) ProtoJSON(code int, i interface{}) error {
	acceptEncoding := c.request.Header().Get(HeaderAcceptEncoding)
	if strings.Contains(acceptEncoding, MIMEApplicationProtobuf) {
		var ok bool
		var m proto.Message
		if m, ok = i.(proto.Message); !ok {
			c.response.Header().Set(HeaderHRPCErr, "true")
			m = statusMSDefault
		}
		c.response.Header().Set(HeaderContentType, MIMEApplicationProtobuf)
		c.response.WriteHeader(code)
		bs, _ := proto.Marshal(m)
		_, err := c.response.Write(bs)
		return err
	}
	c.response.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(i)
}

// XML sends an XML response with status code.
func (c *Context) XML(code int, i interface{}) error {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	}
	//return c.Render(code, xmlRender{Indented: false, Data: i})
	c.response.SetStatus(code)
	c.response.SetData(i)
	pretty := c.QueryBool("_pretty", false)
	c.response.SetEncoder(codec.NewXMLCodec(codec.Indent(pretty)))
	return nil
}

// HTML sends an HTTP response with status code.
func (c *Context) HTML(code int, tmpl *template.Template, name string, obj interface{}) (err error) {
	if c.response.Header().Get(HeaderContentType) == "" {
		c.response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	}
	c.response.SetStatus(code)
	c.response.SetData(obj)
	c.response.SetEncoder(codec.NewHTMLEncoder(tmpl, name))
	return nil
}

type statusErr struct {
	s *rstatus.Status
}

func (e *statusErr) Error() string {
	return fmt.Sprintf("%d:%s", e.s.Code, e.s.Message)
}
func (e *statusErr) Proto() *rstatus.Status {
	if e.s == nil {
		return nil
	}
	return proto.Clone(e.s).(*rstatus.Status)
}

func statusFromString(s string) (*statusErr, bool) {
	i := strings.Index(s, ":")
	if i == -1 {
		return nil, false
	}
	u64, err := strconv.ParseInt(s[:i], 10, 32)
	if err != nil {
		return nil, false
	}

	return &statusErr{
		&rstatus.Status{
			Code:    int32(u64),
			Message: s[i:],
			Details: []*any.Any{},
		},
	}, true
}

var statusMSDefault *rstatus.Status

func init() {
	s, _ := status.FromError(errMicroDefault)
	de, _ := statusFromString(s.Message())
	statusMSDefault = de.Proto()
}
