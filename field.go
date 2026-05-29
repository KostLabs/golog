package golog

import "strconv"

// Field is a pre-typed key/value pair that can be logged without a map
// allocation. Use the constructor helpers (Str, Int, Uint, Float64, Bool,
// etc.) to build fields and pass them to Info / Warn / Error / Debug.
//
// This API is optional and additive — the existing map[string]any API is
// unchanged. Use Field when you need a lower-allocation hot path.
type Field struct {
	key     string
	strVal  string
	intVal  int64
	uintVal uint64
	fltVal  float64
	boolVal bool
	kind    fieldKind
}

type fieldKind uint8

const (
	fieldKindStr fieldKind = iota
	fieldKindInt
	fieldKindUint
	fieldKindFloat
	fieldKindBool
)

// Str creates a string Field.
func Str(key, value string) Field {
	return Field{key: key, strVal: value, kind: fieldKindStr}
}

// Int creates an int Field.
func Int(key string, value int) Field {
	return Field{key: key, intVal: int64(value), kind: fieldKindInt}
}

// Float64 creates a float64 Field.
func Float64(key string, value float64) Field {
	return Field{key: key, fltVal: value, kind: fieldKindFloat}
}

// Bool creates a bool Field.
func Bool(key string, value bool) Field {
	return Field{key: key, boolVal: value, kind: fieldKindBool}
}

// appendFieldBytes encodes a Field directly into dst without allocation.
func appendFieldBytes(dst []byte, f Field) []byte {
	dst = append(dst, ',')
	dst = appendQuoteBytes(dst, f.key)
	dst = append(dst, ':')
	switch f.kind {
	case fieldKindStr:
		dst = appendQuoteBytes(dst, f.strVal)
	case fieldKindInt:
		dst = strconv.AppendInt(dst, f.intVal, 10)
	case fieldKindUint:
		dst = strconv.AppendUint(dst, f.uintVal, 10)
	case fieldKindFloat:
		dst = strconv.AppendFloat(dst, f.fltVal, 'g', -1, 64)
	case fieldKindBool:
		if f.boolVal {
			dst = append(dst, "true"...)
		} else {
			dst = append(dst, "false"...)
		}
	}

	return dst
}
