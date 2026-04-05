package api

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadResponseBodyLimited_AllowsUpToLimit(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Repeat("a", int(DefaultMaxResponseSize)))
	got, err := readResponseBodyLimited(bytes.NewReader(input), DefaultMaxResponseSize)
	if err != nil {
		t.Fatalf("readResponseBodyLimited returned error: %v", err)
	}
	if len(got) != len(input) {
		t.Fatalf("unexpected size: got=%d want=%d", len(got), len(input))
	}
}

func TestReadResponseBodyLimited_RejectsBeyondLimit(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Repeat("a", int(DefaultMaxResponseSize+1)))
	_, err := readResponseBodyLimited(bytes.NewReader(input), DefaultMaxResponseSize)
	if err == nil {
		t.Fatal("expected error for oversized response body")
	}
	if !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}
