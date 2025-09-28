package golog

import (
	"bytes"
	"reflect"
	"strconv"
	"time"
)

// MarshalToBuffer attempts to encode arbitrary values using reflection into
// the provided buffer. It returns an error if it encounters an unsupported
// type (e.g., chan, func, complex) that we don't want to attempt to encode.
func MarshalToBuffer(buf *bytes.Buffer, v any) error {
	return marshalValue(buf, reflect.ValueOf(v))
}

func marshalValue(buf *bytes.Buffer, reflectValue reflect.Value) error {
	if !reflectValue.IsValid() {
		buf.WriteString("null")
		return nil
	}

	for reflectValue.Kind() == reflect.Interface || reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			buf.WriteString("null")
			return nil
		}
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.String:
		buf.WriteString(strconv.Quote(reflectValue.String()))
		return nil
	case reflect.Bool:
		if reflectValue.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(reflectValue.Int(), 10))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(reflectValue.Uint(), 10))
		return nil
	case reflect.Float32, reflect.Float64:
		buf.WriteString(strconv.FormatFloat(reflectValue.Float(), 'g', -1, 64))
		return nil
	case reflect.Map:
		if reflectValue.Type().Key().Kind() != reflect.String {
			return errMarshalTypeUnsupported
		}
		buf.WriteByte('{')
		keys := reflectValue.MapKeys()
		// MapKeys is nondeterministic order; keep original behavior and
		// write in whatever order reflect returns to avoid extra allocs.
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.Quote(k.String()))
			buf.WriteByte(':')
			if err := marshalValue(buf, reflectValue.MapIndex(k)); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	case reflect.Slice, reflect.Array:
		buf.WriteByte('[')
		for i := 0; i < reflectValue.Len(); i++ {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := marshalValue(buf, reflectValue.Index(i)); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	case reflect.Struct:
		if reflectValue.Type() == reflect.TypeOf(time.Time{}) {
			buf.WriteString(strconv.Quote(reflectValue.Interface().(time.Time).UTC().Format(time.RFC3339Nano)))
			return nil
		}
		buf.WriteByte('{')
		reflectionType := reflectValue.Type()
		firstElement := true
		for i := 0; i < reflectionType.NumField(); i++ {
			field := reflectionType.Field(i)
			if field.PkgPath != "" {
				continue
			}
			if !firstElement {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.Quote(field.Name))
			buf.WriteByte(':')
			if err := marshalValue(buf, reflectValue.Field(i)); err != nil {
				return err
			}
			firstElement = false
		}
		buf.WriteByte('}')
		return nil
	default:
		return errMarshalTypeUnsupported
	}
}
