package golog

import "testing"

func TestMapsMerge(t *testing.T) {
	// Given
	dst := map[string]any{"a": 1}

	m1 := map[string]any{"b": 2, "c:": 3}
	m2 := map[string]any{"b": "over", "d": 4}
	var mnil map[string]any = nil

	// When
	mergeMaps(dst, m1, mnil, m2)

	// Then
	// keys should be normalized ("c:" -> "c"), and later maps override earlier ones
	if dst["a"] != 1 {
		t.Fatalf("expected dst[a]=1, got %v", dst["a"])
	}
	if dst["b"] != "over" {
		t.Fatalf("expected dst[b]=\"over\", got %v", dst["b"])
	}
	if dst["c"] != 3 {
		t.Fatalf("expected dst[c]=3, got %v", dst["c"])
	}
	if dst["d"] != 4 {
		t.Fatalf("expected dst[d]=4, got %v", dst["d"])
	}
}
