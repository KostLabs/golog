package golog

import "testing"

func TestFieldConstructorsAndAppendFieldBytes(t *testing.T) {
	tests := []struct {
		name string
		f    Field
		want string
	}{
		{name: "string", f: Str("k", "v"), want: `,"k":"v"`},
		{name: "int", f: Int("n", -7), want: `,"n":-7`},
		{name: "float64", f: Float64("pi", 3.14), want: `,"pi":3.14`},
		{name: "bool true", f: Bool("ok", true), want: `,"ok":true`},
		{name: "bool false", f: Bool("ok", false), want: `,"ok":false`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(appendFieldBytes(nil, tc.f))
			if got != tc.want {
				t.Fatalf("appendFieldBytes mismatch: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestAppendFieldBytesEscapesKeyAndStringValue(t *testing.T) {
	f := Str(`x"y`, "line1\nline2")
	got := string(appendFieldBytes(nil, f))
	want := `,"x\"y":"line1\nline2"`
	if got != want {
		t.Fatalf("escaped field mismatch: got %q want %q", got, want)
	}
}
