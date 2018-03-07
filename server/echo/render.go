package echo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"

	"github.com/ugorji/go/codec"
)

// Render is the interface that wraps the Render function.
type Render interface {
	Render(int, *Context) error
}

// jsonRender defines a JSON render.
type jsonRender struct {
	Data     interface{}
	Indented bool
}

// Render implements Render interface.
func (r jsonRender) Render(code int, c *Context) error {
	encoder := json.NewEncoder(c.response)
	if r.Indented {
		encoder.SetIndent("", "    ")
	}
	c.response.WriteHeader(code)
	return encoder.Encode(r.Data)
}

// stringRender defines a string render.
type stringRender struct {
	Format string
	Data   interface{}
}

// Render implements Render interface.
func (r stringRender) Render(code int, c *Context) error {
	_, err := fmt.Fprintf(c.response, r.Format, r.Data)
	c.response.WriteHeader(code)
	return err
}

// xmlRender defines a XML render.
type xmlRender struct {
	Indented bool
	Data     interface{}
}

// Render implements Render interface.
func (r xmlRender) Render(code int, c *Context) error {
	encoder := xml.NewEncoder(c.response)
	if r.Indented {
		encoder.Indent("", "    ")
	}
	c.response.WriteHeader(code)
	return encoder.Encode(r.Data)
}

type msgpackRender struct {
	Data interface{}
}

// Render implements Render interface.
func (r msgpackRender) Render(code int, c *Context) error {
	var h codec.Handle = new(codec.MsgpackHandle)
	c.response.WriteHeader(code)
	return codec.NewEncoder(c.response, h).Encode(r.Data)
}

// protobufRender defines a protobuf render.
type protobufRender struct {
	Data interface{}
}

// Render implements Render interface.
func (r protobufRender) Render(code int, c *Context) error {
	c.response.WriteHeader(code)
	return nil
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRender struct {
	Templates *template.Template
	Data      interface{}
}

// Render renders a template document
func (t *TemplateRender) Render(code int, c *Context) error {
	c.response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	c.response.WriteHeader(code)

	ret := t.Templates.Execute(c.response, t.Data)
	return ret
}
