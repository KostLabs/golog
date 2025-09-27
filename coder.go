package golog

// coder is responsible for Encoding and Decoding common JSON types

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"time"
)

// FastEncode attempts to write v as JSON into buf using a fast, reflection-free
// path for common primitive types, maps of string->any, and simple slices.
// It returns true when encoding succeeded, or false when the value contains
// a type this fast encoder doesn't support (caller should fall back to
// encoding/json in that case).
func FastEncode(buf *bytes.Buffer, v any) bool {
	return encodeValue(buf, v)
}

func encodeValue(buf *bytes.Buffer, v any) bool {
	switch x := v.(type) {
	case nil:
		buf.WriteString("null")
		return true
	case string:
		// use AppendQuote to avoid an intermediate string where possible
		buf.WriteString(strconv.Quote(x))
		return true
	case bool:
		if x {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return true
	case int, int8, int16, int32:
		buf.WriteString(strconv.FormatInt(int64(x.(int)), 10))
		return true
	case int64:
		buf.WriteString(strconv.FormatInt(x, 10))
		return true
	case uint, uint8, uint16, uint32:
		buf.WriteString(strconv.FormatUint(uint64(x.(uint)), 10))
		return true
	case uint64:
		buf.WriteString(strconv.FormatUint(x, 10))
		return true
	case float32:
		buf.WriteString(strconv.FormatFloat(float64(x), 'g', -1, 32))
		return true
	case float64:
		buf.WriteString(strconv.FormatFloat(x, 'g', -1, 64))
		return true
	case time.Time:
		buf.WriteString(strconv.Quote(x.UTC().Format(time.RFC3339Nano)))
		return true
	case map[string]any:
		return encodeMap(buf, x)
	case []any:
		return encodeSliceAny(buf, x)
	default:
		return false
	}
}

func encodeMap(buf *bytes.Buffer, m map[string]any) bool {
	buf.WriteByte('{')
	first := true
	for k, v := range m {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		fastQuote(buf, k)
		buf.WriteByte(':')

		// Inline the common scalar handling to reduce a function call and
		// keep the fast-path tight. If value has an unsupported type we
		// immediately signal failure so caller can fall back to
		// encoding/json.
		switch w := v.(type) {
		case nil:
			buf.WriteString("null")
		case string:
			fastQuote(buf, w)
		case bool:
			if w {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		case int:
			buf.WriteString(strconv.FormatInt(int64(w), 10))
		case int64:
			buf.WriteString(strconv.FormatInt(w, 10))
		case float64:
			buf.WriteString(strconv.FormatFloat(w, 'g', -1, 64))
		case time.Time:
			buf.WriteString(strconv.Quote(w.UTC().Format(time.RFC3339Nano)))
		case map[string]any:
			if !encodeMap(buf, w) {
				return false
			}
		case []any:
			if !encodeSliceAny(buf, w) {
				return false
			}
		default:
			return false
		}
	}
	buf.WriteByte('}')
	return true
}

func encodeSliceAny(buf *bytes.Buffer, s []any) bool {
	buf.WriteByte('[')
	for i, v := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		if !encodeValue(buf, v) {
			return false
		}
	}
	buf.WriteByte(']')
	return true
}

// MarshalToBuffer attempts to encode arbitrary values using reflection into
// the provided buffer. It returns an error if it encounters an unsupported
// type (e.g., chan, func, complex) that we don't want to attempt to encode.
func MarshalToBuffer(buf *bytes.Buffer, v any) error {
	return marshalValue(buf, reflect.ValueOf(v))
}

var errUnsupported = errors.New("unsupported type for marshal")

func marshalValue(buf *bytes.Buffer, rv reflect.Value) error {
	if !rv.IsValid() {
		buf.WriteString("null")
		return nil
	}

	// If it's an interface or pointer, unwrap
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			buf.WriteString("null")
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		fastQuote(buf, rv.String())
		return nil
	case reflect.Bool:
		if rv.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(rv.Int(), 10))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return nil
	case reflect.Float32, reflect.Float64:
		buf.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, 64))
		return nil
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return errUnsupported
		}
		buf.WriteByte('{')
		keys := rv.MapKeys()
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.Quote(k.String()))
			buf.WriteByte(':')
			if err := marshalValue(buf, rv.MapIndex(k)); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	case reflect.Slice, reflect.Array:
		buf.WriteByte('[')
		l := rv.Len()
		for i := 0; i < l; i++ {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := marshalValue(buf, rv.Index(i)); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	case reflect.Struct:
		// special-case time.Time
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			t := rv.Interface().(time.Time)
			buf.WriteString(strconv.Quote(t.UTC().Format(time.RFC3339Nano)))
			return nil
		}
		// encode exported fields only
		buf.WriteByte('{')
		rt := rv.Type()
		first := true
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}
			if !first {
				buf.WriteByte(',')
			}
			first = false
			name := f.Name
			// honor simple `json:"name"` tag if present
			if tag := f.Tag.Get("json"); tag != "" {
				// take first segment before comma
				if tag == "-" {
					continue
				}
				if comma := indexComma(tag); comma >= 0 {
					name = tag[:comma]
				} else {
					name = tag
				}
			}
			buf.WriteString(strconv.Quote(name))
			buf.WriteByte(':')
			if err := marshalValue(buf, rv.Field(i)); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	default:
		return errUnsupported
	}
}

// indexComma returns the index of first comma or -1
func indexComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

// fastQuote writes a quoted JSON string into buf without allocating a new
// string. It handles the common escapes (\", \\, \n, \r, \t) and writes
// control bytes as \u00XX sequences.
func fastQuote(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if c < 0x20 {
				// control character, write as \u00XX
				const hex = "0123456789abcdef"
				buf.WriteString("\\u00")
				buf.WriteByte(hex[c>>4])
				buf.WriteByte(hex[c&0xF])
			} else {
				buf.WriteByte(c)
			}
		}
	}
	buf.WriteByte('"')
}
