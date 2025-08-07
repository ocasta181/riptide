package checksum

import (
	"bytes"
	"testing"
)

func TestCompute128Deterministic(t *testing.T) {
	data := []byte("riptide")
	s1 := Compute128(data)
	s2 := Compute128(data)
	if !Equal(s1, s2) {
		t.Fatalf("expected equal checksums")
	}
	data2 := []byte("riptide2")
	if Equal(Compute128(data), Compute128(data2)) {
		t.Fatalf("expected different checksums")
	}
}

func TestUint64PairRoundTrip(t *testing.T) {
	data := []byte("roundtrip")
	s := Compute128(data)
	hi, lo := ToUint64Pair(s)
	s2 := FromUint64Pair(hi, lo)
	if !Equal(s, s2) {
		t.Fatalf("expected round trip equality")
	}
	var b1, b2 [16]byte
	copy(b1[:], s[:])
	copy(b2[:], s2[:])
	if !bytes.Equal(b1[:], b2[:]) {
		t.Fatalf("expected byte equality")
	}
}
