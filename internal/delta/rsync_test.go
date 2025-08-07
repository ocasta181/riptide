package delta

import (
	"bytes"
	"testing"
)

func TestComputeFileSigAndDelta_Identity(t *testing.T) {
	basis := []byte("The quick brown fox jumps over the lazy dog.")
	block := 8
	sig := ComputeFileSig(basis, block)
	d := ComputeDelta(sig, basis)
	out := ApplyDelta(basis, d)
	if !bytes.Equal(out, basis) {
		t.Fatalf("identity delta failed: %q != %q", out, basis)
	}
	hasCopy := false
	for _, ins := range d {
		if ins.Op == OpCopy {
			hasCopy = true
			break
		}
	}
	if !hasCopy {
		t.Fatalf("expected at least one copy op")
	}
}

func TestComputeFileSigAndDelta_ModifiedBlock(t *testing.T) {
	basis := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	block := 10
	sig := ComputeFileSig(basis, block)

	newData := make([]byte, len(basis))
	copy(newData, basis)
	start := 20
	for i := 0; i < block && start+i < len(newData); i++ {
		newData[start+i] ^= 0x5A
	}

	d := ComputeDelta(sig, newData)
	out := ApplyDelta(basis, d)
	if !bytes.Equal(out, newData) {
		t.Fatalf("delta apply mismatch")
	}

	hasLit := false
	hasCopy := false
	for _, ins := range d {
		if ins.Op == OpLiteral {
			hasLit = true
		} else if ins.Op == OpCopy {
			hasCopy = true
		}
	}
	if !hasLit || !hasCopy {
		t.Fatalf("expected mix of literal and copy ops, got: %+v", d)
	}
}

func TestComputeFileSigAndDelta_TrailingShortBlock(t *testing.T) {
	basis := []byte("1234567890ABCDEFG")
	block := 7
	sig := ComputeFileSig(basis, block)
	d := ComputeDelta(sig, basis)
	out := ApplyDelta(basis, d)
	if !bytes.Equal(out, basis) {
		t.Fatalf("trailing block identity failed")
	}
}
