package echo

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang/protobuf/proto"
)

// Binder is the interface that wraps the Bind method.
type Binder interface {
	MIME() string
	Bind(*Request, interface{}) error
}

var (
	// JSONBinder defines a json binder.
	JSONBinder = jsonBinder{}
	// XMLBinder defines a XML binder.
	XMLBinder = xmlBinder{}
	// FormBinder defines a form binder.
	FormBinder = formBinder{}
	// QueryBinder defines a query binder.
	QueryBinder = queryBinder{}
	// FormPostBinder defines a form post binder.
	FormPostBinder = formPostBinder{}
	// FormMultipartBinder defines a form multipart binder.
	FormMultipartBinder = formMultipartBinder{}
	// ProtoBufBinder defines a protobuf binder.
	ProtoBufBinder = protobufBinder{}
)

func defaultBinder(method, contentType string) Binder {
	if strings.ToUpper(method) == "GET" {
		return QueryBinder
	}

	contentType = strings.SplitN(contentType, ";", 2)[0]

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
}

type queryBinder struct{}

func (b queryBinder) MIME() string {
	return "query"
}

func (b queryBinder) Bind(req *Request, obj interface{}) error {
	values := req.URL.QueryAll()
	if err := mapForm(obj, values); err != nil {
		return err
	}

	return validate(obj)
}

type jsonBinder struct {
	EnableDecoderUseNumber bool
}

func (b jsonBinder) MIME() string {
	return "json"
}

func (b jsonBinder) Bind(req *Request, obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	if b.EnableDecoderUseNumber {
		decoder.UseNumber()
	}

	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

type xmlBinder struct{}

func (b xmlBinder) MIME() string {
	return "xml"
}

func (b xmlBinder) Bind(req *Request, obj interface{}) error {
	decoder := xml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

type formBinder struct{}
type formPostBinder struct{}
type formMultipartBinder struct{}

func (formBinder) MIME() string {
	return "form"
}

func (formBinder) Bind(req *Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	req.ParseMultipartForm(32 << 10) // 32 MB
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	return validate(obj)
}

func (formPostBinder) MIME() string {
	return "form-urlencoded"
}

func (formPostBinder) Bind(req *Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := mapForm(obj, req.PostForm); err != nil {
		return err
	}
	return validate(obj)
}

func (formMultipartBinder) MIME() string {
	return "multipart/form-data"
}

func (formMultipartBinder) Bind(req *Request, obj interface{}) error {
	if err := req.ParseMultipartForm(32 << 10); err != nil {
		return err
	}
	if err := mapForm(obj, req.MultipartForm.Value); err != nil {
		return err
	}
	return validate(obj)
}

type protobufBinder struct{}

func (protobufBinder) MIME() string {
	return "protobuf"
}

func (protobufBinder) Bind(req *Request, obj interface{}) error {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if err = proto.Unmarshal(buf, obj.(proto.Message)); err != nil {
		return err
	}
	return validate(obj)
}

func validate(obj interface{}) error {
	_, err := govalidator.ValidateStruct(obj)
	return err
}

func mapForm(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = typeField.Name

			// if "form" tag is nil, we inspect if the field is a struct.
			// this would not make sense for JSON parsing but it does for a form
			// since data is flatten
			if structFieldKind == reflect.Struct {
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if _, isTime := structField.Interface().(time.Time); isTime {
				if err := setTimeField(inputValue[0], typeField, structField); err != nil {
					return err
				}
				continue
			}
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return nil
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		return errors.New("Blank time format")
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

// Don't pass in pointers to bind to. Can lead to bugs. See:
// https://github.com/codegangsta/martini-contrib/issues/40
// https://github.com/codegangsta/martini-contrib/pull/34#issuecomment-29683659
func ensureNotPointer(obj interface{}) {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		panic("Pointers are not accepted as binding models")
	}
}
