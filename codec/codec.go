package codec

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"

	"github.com/ugorji/go/codec"
)

type Codec interface {
	//Encode(io.Writer, interface{}) error
	//Decode(io.Reader, interface{}) error
	Encoder
	Decoder
}

type Encoder interface {
	Encode(io.Writer, interface{}) error
}

type Decoder interface {
	Decode(io.Reader, interface{}) error
}

type JSONCodec struct {
	indent    bool
	useNumber bool
}

func NewJSONCodec(opts ...Option) *JSONCodec {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &JSONCodec{
		indent:    options.indent,
		useNumber: options.useNumber,
	}
}

func (c *JSONCodec) Encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	if c.indent {
		encoder.SetIndent("", "    ")
	}
	return encoder.Encode(v)
}

func (c *JSONCodec) Decode(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	if c.useNumber {
		decoder.UseNumber()
	}
	return decoder.Decode(v)
}

type XMLCodec struct {
}

func NewXMLCodec(opts ...Option) *XMLCodec {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	return &XMLCodec{}
}

func (c *XMLCodec) Encode(w io.Writer, v interface{}) error {
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}

func (c *XMLCodec) Decode(r io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(r)
	return decoder.Decode(v)
}

type StringCodec struct {
	format string
}

func NewStringCodec(opts ...Option) *StringCodec {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &StringCodec{
		format: options.format,
	}
}

func (c *StringCodec) Encode(w io.Writer, v interface{}) error {
	var err error
	if c.format == "" {
		_, err = fmt.Fprint(w, v)
	} else {
		_, err = fmt.Fprintf(w, c.format, v)
	}
	return err
}
func (c *StringCodec) Decode(r io.Reader, v interface{}) error {
	return nil
}

type MsgpackCodec struct {
	w io.Writer
	h codec.Handle
}

func NewMsgpackCodec() *MsgpackCodec {
	return &MsgpackCodec{
		w: ioutil.Discard,
		h: new(codec.MsgpackHandle),
	}
}

func (c *MsgpackCodec) Encode(w io.Writer, v interface{}) error {
	// TODO encoder pool
	return codec.NewEncoder(w, c.h).Encode(v)
}

func (c *MsgpackCodec) Decode(r io.Reader, v interface{}) error {
	// TODO decoder pool
	return codec.NewDecoder(r, c.h).Decode(v)
}

type ProtobufCodec struct {
}

func NewPBCodec() *XMLCodec {
	return &XMLCodec{}
}

func (c *ProtobufCodec) Encode(w io.Writer, v interface{}) error {
	return nil
}
func (c *ProtobufCodec) Decode(r io.Reader, v interface{}) error {
	return nil
}

type HTMLEncoder struct {
	tmpl *template.Template
	name string
}

func NewHTMLEncoder(tmpl *template.Template, name string) *HTMLEncoder {
	return &HTMLEncoder{
		tmpl: tmpl,
		name: name,
	}
}

func (c *HTMLEncoder) Encode(w io.Writer, v interface{}) error {
	if len(c.name) == 0 {
		return c.tmpl.Execute(w, v)
	}
	return c.tmpl.ExecuteTemplate(w, c.name, v)
}
