package echo

import (
	"net/http"
	"strconv"
)

// Request wraps an http.Request and implements its interface to be used
// by an HTTP handler to construct an HTTP request.
// See: https://golang.org/pkg/net/http/#Request
type Request struct {
	*http.Request
	*URL
}

func newRequest(req *http.Request) *Request {
	return &Request{
		Request: req,
		URL:     newURL(req.URL),
	}
}

func (r *Request) reset(req *http.Request) {
	r.Request = req
	r.URL = newURL(req.URL)
}

// IsTLS return request TLS flag.
func (r *Request) IsTLS() bool {
	return r.Request.TLS != nil
}

// Header returns request header.
func (r *Request) Header() http.Header {
	return r.Request.Header
}

// QueryString returns request string query.
func (r *Request) QueryString(key string, expect string) string {
	val := r.Query(key)
	if val == "" {
		val = expect
	}
	return val
}

// QueryUint returns request uint query.
func (r *Request) QueryUint(name string, expect uint64) uint64 {
	if res, err := strconv.ParseUint(r.Query(name), 10, 64); err == nil {
		return res
	}
	return expect
}

// QueryInt returns request int query.
func (r *Request) QueryInt(name string, expect int64) int64 {
	if res, err := strconv.ParseInt(r.Query(name), 10, 64); err == nil {
		return res
	}
	return expect
}

// QueryBool returns request boolean query.
func (r *Request) QueryBool(name string, expect bool) bool {
	if res, err := strconv.ParseBool(r.Query(name)); err == nil {
		return res
	}
	return expect
}

// QueryFloat64 returns request float64 query.
func (r *Request) QueryFloat64(name string, expect float64) float64 {
	if res, err := strconv.ParseFloat(r.Query(name), 64); err == nil {
		return res
	}
	return expect
}
