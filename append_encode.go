package golog

import (
	"strconv"
	"time"
)

func appendQuoteBytes(dst []byte, inputString string) []byte {
	dst = append(dst, '"')
	segmentStart := 0
	for charIndex := 0; charIndex < len(inputString); charIndex++ {
		currentChar := inputString[charIndex]
		if currentChar >= 0x20 && currentChar != '\\' && currentChar != '"' {
			continue
		}

		if segmentStart < charIndex {
			dst = append(dst, inputString[segmentStart:charIndex]...)
		}

		switch currentChar {
		case '\\':
			dst = append(dst, `\\`...)
		case '"':
			dst = append(dst, `\"`...)
		case '\n':
			dst = append(dst, `\n`...)
		case '\r':
			dst = append(dst, `\r`...)
		case '\t':
			dst = append(dst, `\t`...)
		default:
			const hexDigits = "0123456789abcdef"
			dst = append(dst, "\\u00"...)
			dst = append(dst, hexDigits[currentChar>>4], hexDigits[currentChar&0xF])
		}

		segmentStart = charIndex + 1
	}

	if segmentStart < len(inputString) {
		dst = append(dst, inputString[segmentStart:]...)
	}

	return append(dst, '"')
}

func appendValueBytes(dst []byte, value any) ([]byte, bool) {
	switch typedValue := value.(type) {
	case nil:
		return append(dst, "null"...), true
	case string:
		return appendQuoteBytes(dst, typedValue), true
	case bool:
		if typedValue {
			return append(dst, "true"...), true
		}
		return append(dst, "false"...), true
	case int:
		return strconv.AppendInt(dst, int64(typedValue), 10), true
	case int8:
		return strconv.AppendInt(dst, int64(typedValue), 10), true
	case int16:
		return strconv.AppendInt(dst, int64(typedValue), 10), true
	case int32:
		return strconv.AppendInt(dst, int64(typedValue), 10), true
	case int64:
		return strconv.AppendInt(dst, typedValue, 10), true
	case uint:
		return strconv.AppendUint(dst, uint64(typedValue), 10), true
	case uint8:
		return strconv.AppendUint(dst, uint64(typedValue), 10), true
	case uint16:
		return strconv.AppendUint(dst, uint64(typedValue), 10), true
	case uint32:
		return strconv.AppendUint(dst, uint64(typedValue), 10), true
	case uint64:
		return strconv.AppendUint(dst, typedValue, 10), true
	case float32:
		return strconv.AppendFloat(dst, float64(typedValue), 'g', -1, 32), true
	case float64:
		return strconv.AppendFloat(dst, typedValue, 'g', -1, 64), true
	case time.Time:
		dst = append(dst, '"')
		t := typedValue.UTC()
		var tsBuf [64]byte
		dst = append(dst, appendRFC3339NanoUTC(tsBuf[:0], t)...)
		dst = append(dst, '"')
		return dst, true
	case map[string]any:
		return appendMapBytes(dst, typedValue)
	case []any:
		return appendSliceBytes(dst, typedValue)
	default:
		return dst, false
	}
}

func appendMapBytes(dst []byte, mapData map[string]any) ([]byte, bool) {
	dst = append(dst, '{')
	first := true
	for key, value := range mapData {
		if !first {
			dst = append(dst, ',')
		}
		first = false
		dst = appendQuoteBytes(dst, key)
		dst = append(dst, ':')
		var ok bool
		dst, ok = appendValueBytes(dst, value)
		if !ok {
			return dst, false
		}
	}
	dst = append(dst, '}')
	return dst, true
}

func appendSliceBytes(dst []byte, values []any) ([]byte, bool) {
	dst = append(dst, '[')
	for i, value := range values {
		if i > 0 {
			dst = append(dst, ',')
		}
		var ok bool
		dst, ok = appendValueBytes(dst, value)
		if !ok {
			return dst, false
		}
	}
	dst = append(dst, ']')
	return dst, true
}
