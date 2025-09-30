package golog

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestFastEncodePrimitives(t *testing.T) {
	var buf bytes.Buffer

	if !FastEncode(&buf, nil) {
		t.Fatalf("FastEncode(nil) returned false")
	}
	if buf.String() != "null" {
		t.Fatalf("expected null, got %q", buf.String())
	}

	buf.Reset()
	s := "a\n\tb\"c\\d"
	if !FastEncode(&buf, s) {
		t.Fatalf("FastEncode(string) returned false")
	}
	if buf.String() != strconv.Quote(s) {
		t.Fatalf("quoted string mismatch: expected %s got %s", strconv.Quote(s), buf.String())
	}

	buf.Reset()
	tm := time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)
	if !FastEncode(&buf, tm) {
		t.Fatalf("FastEncode(time.Time) returned false")
	}
	var gotTime string
	if err := json.Unmarshal(buf.Bytes(), &gotTime); err != nil {
		t.Fatalf("unmarshal encoded time: %v", err)
	}
	if gotTime != tm.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("time formatting mismatch: expected %q got %q", tm.UTC().Format(time.RFC3339Nano), gotTime)
	}
}

func TestFastEncodeMapAndSlice(t *testing.T) {
	data := map[string]any{
		"a": "x",
		"b": 123,
		"c": []any{1, "s", map[string]any{"z": true}},
	}

	var buf bytes.Buffer
	if !FastEncode(&buf, data) {
		t.Fatalf("FastEncode(map) returned false")
	}

	var got any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal fast-encoded map: %v", err)
	}

	expectedBytes, _ := json.Marshal(data)
	var want any
	if err := json.Unmarshal(expectedBytes, &want); err != nil {
		t.Fatalf("unmarshal json.Marshal(data): %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("encoded map/slice mismatch: got %#v want %#v", got, want)
	}
}

func TestFastEncodeUnsupported(t *testing.T) {
	var buf bytes.Buffer
	ch := make(chan int)
	if FastEncode(&buf, ch) {
		t.Fatalf("expected FastEncode to return false for chan")
	}

	// a struct value that's not one of the supported fast types should fail
	type S struct{ X int }
	buf.Reset()
	if FastEncode(&buf, S{X: 1}) {
		t.Fatalf("expected FastEncode to return false for struct S")
	}
}

func TestFastQuoteControlCharsAndNumericEdgecases(t *testing.T) {
	// control characters should be encoded and roundtripped by json.Unmarshal
	s := string([]byte{0x01, 0x02, 'A', '\n'})
	var buf bytes.Buffer
	if !FastEncode(&buf, s) {
		t.Fatalf("FastEncode(control string) returned false")
	}
	var out string
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal quoted control string: %v", err)
	}
	if out != s {
		t.Fatalf("control string mismatch: expected %#v got %#v", s, out)
	}

	// map with supported numeric types at top-level via encodeValue
	var buf2 bytes.Buffer
	if !FastEncode(&buf2, int32(-5)) {
		t.Fatalf("FastEncode(int32) returned false")
	}
	if buf2.String() != "-5" {
		t.Fatalf("int32 encoded mismatch: got %s", buf2.String())
	}

	// but map[string]any with a uint value is not supported by inline switch and should fail
	m := map[string]any{"u": uint(7)}
	var buf3 bytes.Buffer
	if FastEncode(&buf3, m) {
		t.Fatalf("expected FastEncode(map[string]any with uint) to return false")
	}
}
