package pipeline

import (
	"sort"
	"testing"

	"riptide/internal/cryptoutil"
	"riptide/internal/queue"
)

func TestSenderReceiverPipelineE2E(t *testing.T) {
	var key [32]byte
	for i := range key {
		key[i] = byte(i + 1)
	}
	a, err := cryptoutil.NewAEAD(key)
	if err != nil {
		t.Fatalf("aead: %v", err)
	}
	enc := &Encryptor{AEAD: a, AAD: []byte("ctx")}

	src := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i * 31)
	}
	chunkSize := 128
	ds := Chunk(src, chunkSize)

	withSums, err := ApplyTransforms(ds, ComputeChecksum())
	if err != nil {
		t.Fatalf("checksums: %v", err)
	}
	encrypted, err := ApplyTransforms(withSums, Encrypt(enc))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	r := queue.NewRing[Descriptor](len(encrypted) + 8)
	for _, d := range encrypted {
		if !r.Enqueue(d) {
			t.Fatalf("enqueue failed")
		}
	}

	var netDrain []Descriptor
	for {
		d, ok := r.Dequeue()
		if !ok {
			break
		}
		netDrain = append(netDrain, d)
	}

	decrypted, err := ApplyTransforms(netDrain, Decrypt(enc))
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	verified, err := ApplyTransforms(decrypted, VerifyChecksum())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	sort.Slice(verified, func(i, j int) bool { return verified[i].Offset < verified[j].Offset })
	out := make([]byte, 0, len(src))
	for _, d := range verified {
		out = append(out, d.Data...)
	}
	if len(out) != len(src) {
		t.Fatalf("length mismatch: %d != %d", len(out), len(src))
	}
	for i := range out {
		if out[i] != src[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}
