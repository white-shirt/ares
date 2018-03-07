package echo

import (
	"io"
	"net/http"

	"github.com/sevenNt/ares/codec"
)

// Response wraps an http.ResponseWriter and implements its interface to be used
// by an HTTP handler to construct an HTTP response.
// See: https://golang.org/pkg/net/http/#ResponseWriter
type Response struct {
	http.ResponseWriter
	status int
	size   int64
	writer io.Writer

	data    interface{}
	encoder codec.Encoder

	commited bool
}

func (r *Response) SetWriter(w io.Writer) {
	r.writer = w
}

func (r *Response) Writer() io.Writer {
	return r.writer
}

func (r *Response) SetData(v interface{}) {
	r.data = v
}

func (r *Response) Data() interface{} {
	return r.data
}

func (r *Response) SetEncoder(encoder codec.Encoder) {
	r.encoder = encoder
}

func (r *Response) Encoder() codec.Encoder {
	return r.encoder
}

func newResponse(w http.ResponseWriter) (r *Response) {
	return &Response{
		status:         StatusOK,
		ResponseWriter: w,
		writer:         w,
	}
}

func (r *Response) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = StatusOK
	r.writer = w
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly
// used to send error codes.
func (r *Response) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	r.commited = true
}

// Status returns response status.
func (r *Response) Status() int {
	return r.status
}

func (r *Response) SetStatus(status int) {
	r.status = status
}

// Write writes the data to the connection as part of an HTTP reply.
//func (r *Response) Write(bs []byte) (n int, e error) {
//n, e = r.ResponseWriter.Write(bs)
//fmt.Println("response.write: ", string(bs))
//r.size += int64(n)
//return
//}

// Size returns response size.
func (r *Response) Size() int64 {
	return r.size
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	if r.status == StatusNoContent {
		r.ResponseWriter.Header().Del(HeaderContentEncoding)
	}

	if !r.commited {
		r.ResponseWriter.WriteHeader(r.status)
	}

	if r.encoder != nil {
		r.encoder.Encode(r.writer, r.data)
	}
}
