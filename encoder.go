package golog

// coder is responsible for Encoding and Decoding common JSON types

import (
	"bytes"
	"strconv"
	"time"
)

// FastEncode attempts to write value as JSON into buffer using a fast, reflection-free
// path for common primitive types, maps of string->any, and simple slices.
// It returns true when encoding succeeded, or false when the value contains
// a type this fast encoder doesn't support (caller should fall back to
// encoding/json in that case).
func FastEncode(buffer *bytes.Buffer, value any) bool {
	return encodeValue(buffer, value)
}

func encodeValue(buffer *bytes.Buffer, value any) bool {
	switch typedValue := value.(type) {
	case nil:
		buffer.WriteString("null")
		return true
	case string:
		fastQuote(buffer, typedValue)
		return true
	case bool:
		if typedValue {
			buffer.WriteString("true")
		} else {
			buffer.WriteString("false")
		}
		return true
	case int:
		fastFormatInt(buffer, int64(typedValue))
		return true
	case int8:
		fastFormatInt(buffer, int64(typedValue))
		return true
	case int16:
		fastFormatInt(buffer, int64(typedValue))
		return true
	case int32:
		fastFormatInt(buffer, int64(typedValue))
		return true
	case int64:
		fastFormatInt(buffer, typedValue)
		return true
	case uint:
		fastFormatUint(buffer, uint64(typedValue))
		return true
	case uint8:
		fastFormatUint(buffer, uint64(typedValue))
		return true
	case uint16:
		fastFormatUint(buffer, uint64(typedValue))
		return true
	case uint32:
		fastFormatUint(buffer, uint64(typedValue))
		return true
	case uint64:
		fastFormatUint(buffer, typedValue)
		return true
	case float32:
		buffer.WriteString(strconv.FormatFloat(float64(typedValue), 'g', -1, 32))
		return true
	case float64:
		buffer.WriteString(strconv.FormatFloat(typedValue, 'g', -1, 64))
		return true
	case time.Time:
		buffer.WriteByte('"')
		buffer.WriteString(typedValue.UTC().Format(time.RFC3339Nano))
		buffer.WriteByte('"')
		return true
	case map[string]any:
		return encodeMap(buffer, typedValue)
	case []any:
		return encodeSliceAny(buffer, typedValue)
	default:
		return false
	}
}

func encodeMap(buffer *bytes.Buffer, mapData map[string]any) bool {
	buffer.WriteByte('{')
	isFirstField := true
	for key, value := range mapData {
		if !isFirstField {
			buffer.WriteByte(',')
		}
		isFirstField = false
		fastQuote(buffer, key)
		buffer.WriteByte(':')

		// Inline the common scalar handling to reduce a function call and
		// keep the fast-path tight. If value has an unsupported type we
		// immediately signal failure so caller can fall back to
		// encoding/json.
		switch typedValue := value.(type) {
		case nil:
			buffer.WriteString("null")
		case string:
			fastQuote(buffer, typedValue)
		case bool:
			if typedValue {
				buffer.WriteString("true")
			} else {
				buffer.WriteString("false")
			}
		case int:
			fastFormatInt(buffer, int64(typedValue))
		case int64:
			fastFormatInt(buffer, typedValue)
		case float64:
			buffer.WriteString(strconv.FormatFloat(typedValue, 'g', -1, 64))
		case time.Time:
			buffer.WriteByte('"')
			buffer.WriteString(typedValue.UTC().Format(time.RFC3339Nano))
			buffer.WriteByte('"')
		case map[string]any:
			if !encodeMap(buffer, typedValue) {
				return false
			}
		case []any:
			if !encodeSliceAny(buffer, typedValue) {
				return false
			}
		default:
			return false
		}
	}
	buffer.WriteByte('}')
	return true
}

func encodeSliceAny(buffer *bytes.Buffer, slice []any) bool {
	buffer.WriteByte('[')
	for index, value := range slice {
		if index > 0 {
			buffer.WriteByte(',')
		}
		if !encodeValue(buffer, value) {
			return false
		}
	}
	buffer.WriteByte(']')
	return true
}

// fastQuote writes a quoted JSON string into buffer without allocating a new
// string. It handles the common escapes (\", \\, \n, \r, \t) and writes
// control bytes as \u00XX sequences. This is used on the hot fast-path to
// avoid extra allocations from strconv.Quote.
func fastQuote(buffer *bytes.Buffer, inputString string) {
	buffer.WriteByte('"')
	for charIndex := 0; charIndex < len(inputString); charIndex++ {
		currentChar := inputString[charIndex]
		switch currentChar {
		case '\\':
			buffer.WriteString(`\\`)
		case '"':
			buffer.WriteString(`\"`)
		case '\n':
			buffer.WriteString(`\n`)
		case '\r':
			buffer.WriteString(`\r`)
		case '\t':
			buffer.WriteString(`\t`)
		default:
			if currentChar < 0x20 {
				// control character, write as \u00XX
				const hexDigits = "0123456789abcdef"
				buffer.WriteString("\\u00")
				buffer.WriteByte(hexDigits[currentChar>>4])
				buffer.WriteByte(hexDigits[currentChar&0xF])
			} else {
				buffer.WriteByte(currentChar)
			}
		}
	}
	buffer.WriteByte('"')
}

// fastFormatInt writes an int64 directly to buffer without string allocation
func fastFormatInt(buffer *bytes.Buffer, integerValue int64) {
	if integerValue == 0 {
		buffer.WriteByte('0')
		return
	}

	var digitBuffer [20]byte
	bufferPosition := len(digitBuffer)
	isNegative := integerValue < 0
	if isNegative {
		integerValue = -integerValue
	}

	for integerValue > 0 {
		bufferPosition--
		digitBuffer[bufferPosition] = '0' + byte(integerValue%10)
		integerValue /= 10
	}

	if isNegative {
		bufferPosition--
		digitBuffer[bufferPosition] = '-'
	}

	buffer.Write(digitBuffer[bufferPosition:])
}

// fastFormatUint writes a uint64 directly to buffer without string allocation
func fastFormatUint(buffer *bytes.Buffer, unsignedValue uint64) {
	if unsignedValue == 0 {
		buffer.WriteByte('0')
		return
	}

	var digitBuffer [20]byte
	bufferPosition := len(digitBuffer)

	for unsignedValue > 0 {
		bufferPosition--
		digitBuffer[bufferPosition] = '0' + byte(unsignedValue%10)
		unsignedValue /= 10
	}

	buffer.Write(digitBuffer[bufferPosition:])
}
