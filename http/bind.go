package http

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	kiterrors "github.com/quocdaitrn/golang-kit/errors"
	"github.com/quocdaitrn/golang-kit/util/objutil"
)

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from an form or query param.
	UnmarshalParam(param string) error
}

// unmarshalFuncs is a map of some types and their bindUnmarshalFunc.
var unmarshalFuncs = map[reflect.Type]bindUnmarshalFunc{
	reflect.TypeOf(primitive.ObjectID{}): unmarshalBsonObjectID,
}

type bindUnmarshalFunc func(value string, field reflect.Value) error

const (
	defaultMemory = 32 << 20 // 32 MB
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Bind binds data from body, header, query and path param to the result by
// priorities: Header > Path Param > Query > Body
func Bind(r *http.Request, res interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(res))
	typ := val.Type()
	names := objutil.FlattenFieldNames(typ)

	if r.ContentLength != 0 && r.Method != http.MethodGet {
		ctyp := r.Header.Get(HeaderContentType)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		switch {
		case strings.HasPrefix(ctyp, MIMEApplicationJSON):
			if err := json.NewDecoder(r.Body).Decode(res); err != nil {
				return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
			}
		case strings.HasPrefix(ctyp, MIMEApplicationXML), strings.HasPrefix(ctyp, MIMETextXML):
			if err := xml.NewDecoder(r.Body).Decode(res); err != nil {
				if ute, ok := err.(*xml.UnsupportedTypeError); ok {
					return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(fmt.Sprintf("Unsupported type error: type=%v, error=%v", ute.Type, ute.Error())))
				} else if se, ok := err.(*xml.SyntaxError); ok {
					return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(fmt.Sprintf("Syntax error: line=%v, error=%v", se.Line, se.Error())))
				} else {
					return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
				}
			}
		case strings.HasPrefix(ctyp, MIMEApplicationForm), strings.HasPrefix(ctyp, MIMEMultipartForm):
			if strings.HasPrefix(r.Header.Get(HeaderContentType), MIMEMultipartForm) {
				if err := r.ParseMultipartForm(defaultMemory); err != nil {
					return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
				}
			} else {
				if err := r.ParseForm(); err != nil {
					return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
				}
			}

			if err := bindData(r.Form, "form", names, true, res); err != nil {
				return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
			}
		default:
			return kiterrors.ErrHTTPUnsupportedMediaType
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Bind query param
	if err := bindData(r.URL.Query(), "query", names, true, res); err != nil {
		return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
	}

	// Bind path param
	if err := bindPathParam(r, names, res); err != nil {
		return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
	}

	// Bind header param
	if err := bindData(r.Header, "header", names, false, res); err != nil {
		return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(err))
	}

	return nil
}

// bindPathParam binds data from path param to the result.
func bindPathParam(r *http.Request, flattenFieldNames map[string]bool, res interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(res))
	typ := val.Type()

	for name := range flattenFieldNames {
		field := val.FieldByName(name)
		fieldTyp, hasField := typ.FieldByName(name)

		if !field.IsValid() || !field.CanSet() || !hasField {
			continue
		}

		fieldKind := field.Kind()

		paramTag := fieldTyp.Tag.Get("param")
		paramTags := strings.Split(paramTag, ",")
		paramName := paramTags[0]

		if paramTag == "-" || paramTag == "" {
			continue
		}

		inputValue := mux.Vars(r)[paramName]
		if inputValue == "" {
			continue
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(fieldKind, inputValue, field); ok {
			if err != nil {
				return err
			}
			continue
		}

		if err := setWithProperType(fieldKind, inputValue, field); err != nil {
			return err
		}

	}
	return nil
}

// bindData binds input data to the result by the input tag.
func bindData(data map[string][]string, tag string, flattenFieldNames map[string]bool, isCaseSensitive bool, res interface{}) error {
	bkData := data
	data = map[string][]string{}
	for k, v := range bkData {
		data[strings.ToLower(k)] = v
	}

	val := reflect.Indirect(reflect.ValueOf(res))
	typ := val.Type()

	for name := range flattenFieldNames {
		field := val.FieldByName(name)
		structFiled, hasField := typ.FieldByName(name)

		if !field.IsValid() || !field.CanSet() || !hasField {
			continue
		}

		fieldKind := field.Kind()

		fieldName := structFiled.Tag.Get(tag)
		if fieldName == "" {
			continue
		}

		var inputValue []string
		if isCaseSensitive {
			exists := false
			inputValue, exists = bkData[fieldName]
			if !exists {
				continue
			}
		} else {
			exists := false
			inputValue, exists = data[strings.ToLower(fieldName)]
			if !exists {
				continue
			}
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(fieldKind, inputValue[0], field); ok {
			if err != nil {
				return err
			}
			continue
		}

		numElem := len(inputValue)
		if fieldKind == reflect.Slice && numElem > 0 {
			sliceOf := field.Type().Elem().Kind()
			slice := reflect.MakeSlice(field.Type(), numElem, numElem)
			for j := 0; j < numElem; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			field.Set(slice)
		} else {
			if err := setWithProperType(fieldKind, inputValue[0], field); err != nil {
				return err
			}
		}
	}
	return nil
}

// setWithProperType sets input values input in string to field by it's type.
func setWithProperType(kind reflect.Kind, val string, field reflect.Value) error {
	// But also call it here, in case we're dealing with an array of BindUnmarshalers
	if ok, err := unmarshalField(kind, val, field); ok {
		return err
	}

	switch kind {
	case reflect.Ptr:
		return setWithProperType(field.Elem().Kind(), val, field.Elem())
	case reflect.Int:
		return setIntField(val, 0, field)
	case reflect.Int8:
		return setIntField(val, 8, field)
	case reflect.Int16:
		return setIntField(val, 16, field)
	case reflect.Int32:
		return setIntField(val, 32, field)
	case reflect.Int64:
		return setIntField(val, 64, field)
	case reflect.Uint:
		return setUintField(val, 0, field)
	case reflect.Uint8:
		return setUintField(val, 8, field)
	case reflect.Uint16:
		return setUintField(val, 16, field)
	case reflect.Uint32:
		return setUintField(val, 32, field)
	case reflect.Uint64:
		return setUintField(val, 64, field)
	case reflect.Bool:
		return setBoolField(val, field)
	case reflect.Float32:
		return setFloatField(val, 32, field)
	case reflect.Float64:
		return setFloatField(val, 64, field)
	case reflect.String:
		field.SetString(val)
	case reflect.Interface:
		field.Set(reflect.ValueOf(val))
	default:
		return fmt.Errorf("unsupported kind: %s", kind.String())
	}
	return nil
}

// setIntField sets input values input in string to a int field.
func setIntField(value string, bitSize int, field reflect.Value) error {
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

// setUintField sets input values input in string to a uint field.
func setUintField(value string, bitSize int, field reflect.Value) error {
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

// setBoolField sets input values input in string to a bool field.
func setBoolField(value string, field reflect.Value) error {
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

// setFloatField sets input values input in string to a float field.
func setFloatField(value string, bitSize int, field reflect.Value) error {
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func getBindUnmarshaler(field reflect.Value) (BindUnmarshaler, bool) {
	ptr := reflect.New(field.Type())
	if ptr.CanInterface() {
		iFace := ptr.Interface()
		if unmarshaler, ok := iFace.(BindUnmarshaler); ok {
			return unmarshaler, ok
		}
	}
	return nil, false
}

func getBindUnmarshalFunc(field reflect.Value) (bindUnmarshalFunc, bool) {
	if unmarshalFunc, ok := unmarshalFuncs[field.Type()]; ok {
		return unmarshalFunc, true
	}
	return nil, false
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	if unmarshaler, ok := getBindUnmarshaler(field); ok {
		err := unmarshaler.UnmarshalParam(value)
		field.Set(reflect.ValueOf(unmarshaler).Elem())
		return true, err
	} else if unmarshalFunc, ok := getBindUnmarshalFunc(field); ok {
		err := unmarshalFunc(value, field)
		return true, err
	}
	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		// Initialize the pointer to a nil value
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}

func unmarshalBsonObjectID(value string, field reflect.Value) error {
	if objectID, err := primitive.ObjectIDFromHex(value); err == nil {
		field.Set(reflect.ValueOf(objectID))
		return nil
	}

	return kiterrors.WithStack(kiterrors.ErrRequestBindingFailed.WithDetails(fmt.Sprintf("custom binder marshaling: cannot unmarshal %v as %s", value, field.Type().Name())))
}
