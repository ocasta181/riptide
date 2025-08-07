package pipeline

import (
	"errors"
	"testing"

	"riptide/internal/checksum"
)

func TestChunk_Basic(t *testing.T) {
	src := []byte("abcdefghij")
	ds := Chunk(src, 3)
	if len(ds) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(ds))
	}
	wantLens := []int{3, 3, 3, 1}
	wantOffsets := []uint64{0, 3, 6, 9}
	for i, d := range ds {
		if len(d.Data) != wantLens[i] {
			t.Fatalf("chunk %d len = %d want %d", i, len(d.Data), wantLens[i])
		}
		if d.Offset != wantOffsets[i] {
			t.Fatalf("chunk %d offset = %d want %d", i, d.Offset, wantOffsets[i])
		}
	}
	src[0] = 'Z'
	if string(ds[0].Data) != "abc" {
		t.Fatalf("expected first chunk to be 'abc', got %q", ds[0].Data)
	}
}

func TestChunk_ZeroOrNegativeSize(t *testing.T) {
	src := []byte("abcd")
	if got := Chunk(src, 0); len(got) != 4 {
		t.Fatalf("chunk(0) expected 4 chunks, got %d", len(got))
	}
	if got := Chunk(src, -5); len(got) != 4 {
		t.Fatalf("chunk(-5) expected 4 chunks, got %d", len(got))
	}
}

func TestApplyTransforms_WithComputeChecksum(t *testing.T) {
	src := []byte("hello world")
	ds := Chunk(src, 5)
	out, err := ApplyTransforms(ds, ComputeChecksum())
	if err != nil {
		t.Fatalf("apply err: %v", err)
	}
	if len(out) != len(ds) {
		t.Fatalf("length mismatch")
	}
	for i := range out {
		sum := checksum.Compute128(out[i].Data)
		if !checksum.Equal(sum, out[i].Sum) {
			t.Fatalf("checksum mismatch on chunk %d", i)
		}
	}
}

func TestApplyTransforms_ErrorPropagation(t *testing.T) {
	ds := []Descriptor{{Data: []byte("x")}}
	errT := errors.New("boom")
	failing := func(d Descriptor) (Descriptor, error) { return d, errT }
	_, err := ApplyTransforms(ds, failing)
	if err == nil || !errors.Is(err, errT) {
		t.Fatalf("expected error propagation, got %v", err)
	}
}
