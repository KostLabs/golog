package golog

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMarshalPrimitivesAndTime(t *testing.T) {
	var buf bytes.Buffer
	if err := MarshalToBuffer(&buf, "hello\n"); err != nil {
		t.Fatalf("MarshalToBuffer(string) error: %v", err)
	}
	var s string
	if err := json.Unmarshal(buf.Bytes(), &s); err != nil {
		t.Fatalf("unmarshal string: %v", err)
	}
	if s != "hello\n" {
		t.Fatalf("string mismatch: got %q", s)
	}

	buf.Reset()
	tm := time.Date(2021, 12, 31, 23, 59, 59, 123456789, time.UTC)
	if err := MarshalToBuffer(&buf, tm); err != nil {
		t.Fatalf("MarshalToBuffer(time) error: %v", err)
	}
	var ts string
	if err := json.Unmarshal(buf.Bytes(), &ts); err != nil {
		t.Fatalf("unmarshal time: %v", err)
	}
	if ts != tm.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("time mismatch: expected %q got %q", tm.UTC().Format(time.RFC3339Nano), ts)
	}
}

func TestMarshalMapSliceStruct(t *testing.T) {
	type Inner struct {
		N int
		S string
	}
	type Outer struct {
		I Inner
		A []int
		M map[string]any
	}

	val := Outer{
		I: Inner{N: 2, S: "ok"},
		A: []int{1, 2, 3},
		M: map[string]any{"k": "v"},
	}

	var buf bytes.Buffer
	if err := MarshalToBuffer(&buf, val); err != nil {
		t.Fatalf("MarshalToBuffer(struct) error: %v", err)
	}

	var got any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal marshaled struct: %v", err)
	}

	// Compare against encoding/json result
	expected, err := json.Marshal(val)
	if err != nil {
		t.Fatalf("json.Marshal(val) error: %v", err)
	}
	var want any
	if err := json.Unmarshal(expected, &want); err != nil {
		t.Fatalf("unmarshal expected json: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marshal output mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMarshalUnsupportedTypes(t *testing.T) {
	var buf bytes.Buffer
	// func is unsupported
	if err := MarshalToBuffer(&buf, func() {}); err != errMarshalTypeUnsupported {
		t.Fatalf("expected errMarshalTypeUnsupported for func, got: %v", err)
	}

	buf.Reset()
	// map with non-string key is unsupported
	m := map[int]string{1: "a"}
	if err := MarshalToBuffer(&buf, m); err != errMarshalTypeUnsupported {
		t.Fatalf("expected errMarshalTypeUnsupported for map[int]string, got: %v", err)
	}
}

func TestMarshalNilPointerAndInterfaceAndArray(t *testing.T) {
	var buf bytes.Buffer

	var p *int = nil
	if err := MarshalToBuffer(&buf, p); err != nil {
		t.Fatalf("expected nil pointer to marshal to null, got error: %v", err)
	}
	if buf.String() != "null" {
		t.Fatalf("expected null for nil pointer, got %s", buf.String())
	}

	buf.Reset()
	var i any = nil
	if err := MarshalToBuffer(&buf, i); err != nil {
		t.Fatalf("expected nil interface to marshal to null, got error: %v", err)
	}
	if buf.String() != "null" {
		t.Fatalf("expected null for nil interface, got %s", buf.String())
	}

	buf.Reset()
	arr := [3]int{1, 2, 3}
	if err := MarshalToBuffer(&buf, arr); err != nil {
		t.Fatalf("expected array to marshal, got error: %v", err)
	}
	var out []int
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal array: %v", err)
	}
	if len(out) != 3 || out[0] != 1 || out[2] != 3 {
		t.Fatalf("unexpected array content: %v", out)
	}

	// pointer to struct
	buf.Reset()
	type P struct{ A int }
	pp := &P{A: 5}
	if err := MarshalToBuffer(&buf, pp); err != nil {
		t.Fatalf("pointer to struct marshal error: %v", err)
	}
	var pm map[string]any
	if err := json.Unmarshal(buf.Bytes(), &pm); err != nil {
		t.Fatalf("unmarshal pointer-to-struct: %v", err)
	}
	if pm["A"] != float64(5) {
		t.Fatalf("expected A=5, got %v", pm["A"])
	}
}
