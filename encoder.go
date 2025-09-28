package golog

// coder is responsible for Encoding and Decoding common JSON types

import (
	"bytes"
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
		fastQuote(buf, x)
		return true
	case bool:
		if x {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return true
	case int:
		buf.WriteString(strconv.FormatInt(int64(x), 10))
		return true
	case int8:
		buf.WriteString(strconv.FormatInt(int64(x), 10))
		return true
	case int16:
		buf.WriteString(strconv.FormatInt(int64(x), 10))
		return true
	case int32:
		buf.WriteString(strconv.FormatInt(int64(x), 10))
		return true
	case int64:
		buf.WriteString(strconv.FormatInt(x, 10))
		return true
	case uint:
		buf.WriteString(strconv.FormatUint(uint64(x), 10))
		return true
	case uint8:
		buf.WriteString(strconv.FormatUint(uint64(x), 10))
		return true
	case uint16:
		buf.WriteString(strconv.FormatUint(uint64(x), 10))
		return true
	case uint32:
		buf.WriteString(strconv.FormatUint(uint64(x), 10))
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

// fastQuote writes a quoted JSON string into buf without allocating a new
// string. It handles the common escapes (\", \\, \n, \r, \t) and writes
// control bytes as \u00XX sequences. This is used on the hot fast-path to
// avoid extra allocations from strconv.Quote.
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
