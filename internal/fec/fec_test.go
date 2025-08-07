package fec

import (
	"bytes"
	"testing"
)

func TestBuildShardsAndReconstruct(t *testing.T) {
	c, err := NewCodec(4, 2)
	if err != nil {
		t.Fatalf("codec: %v", err)
	}
	data := [][]byte{
		[]byte("AAAA"),
		[]byte("BBBB"),
		[]byte("CCCC"),
		[]byte("DDDD"),
	}
	shards, err := c.BuildShards(data)
	if err != nil {
		t.Fatalf("build shards: %v", err)
	}
	if len(shards) != c.DataShards()+c.ParityShards() {
		t.Fatalf("shard count mismatch: %d", len(shards))
	}
	shards[1] = nil
	if err := c.Reconstruct(shards); err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	for i := 0; i < c.DataShards(); i++ {
		if !bytes.Equal(shards[i], data[i]) {
			t.Fatalf("data shard %d mismatch: got %q want %q", i, shards[i], data[i])
		}
	}
}

func TestBuildShardsErrors(t *testing.T) {
	if _, err := NewCodec(0, 2); err == nil {
		t.Fatalf("expected invalid shard counts")
	}
	c, err := NewCodec(2, 1)
	if err != nil {
		t.Fatalf("codec: %v", err)
	}
	// wrong number of data shards
	if _, err := c.BuildShards([][]byte{[]byte("AA")}); err == nil {
		t.Fatalf("expected wrong number of shards error")
	}
	// unequal sizes
	if _, err := c.BuildShards([][]byte{[]byte("AA"), []byte("BBB")}); err == nil {
		t.Fatalf("expected unequal shard sizes error")
	}
}

func TestSelectParity(t *testing.T) {
	if got := SelectParity(0.0, 4); got != 1 {
		t.Fatalf("low loss expect 1, got %d", got)
	}
	if got := SelectParity(0.01, 4); got != 2 {
		t.Fatalf("0.01 expect 2, got %d", got)
	}
	if got := SelectParity(0.03, 4); got != 3 {
		t.Fatalf("0.03 expect 3, got %d", got)
	}
	if got := SelectParity(0.07, 4); got != 4 {
		t.Fatalf("0.07 expect 4, got %d", got)
	}
	if got := SelectParity(0.5, 3); got != 3 {
		t.Fatalf("high loss expect cap 3, got %d", got)
	}
	if got := SelectParity(0.0, 0); got != 0 {
		t.Fatalf("maxParity 0 expect 0, got %d", got)
	}
}
