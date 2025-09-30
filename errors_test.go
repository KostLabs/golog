package golog

import "testing"

func TestErrMarshalTypeUnsupportedIsSentinel(t *testing.T) {
	if errMarshalTypeUnsupported == nil {
		t.Fatalf("errMarshalTypeUnsupported should not be nil")
	}

	// Ensure the sentinel value is comparable and not equal to a different error
	if errMarshalTypeUnsupported == (error)(nil) {
		t.Fatalf("unexpected nil sentinel")
	}

	if errMarshalTypeUnsupported.Error() == "" {
		t.Fatalf("expected non-empty error string")
	}
}
