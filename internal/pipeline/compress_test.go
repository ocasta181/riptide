package pipeline

import (
	"testing"

	"riptide/internal/cryptoutil"
)

func TestCompressDecompressRoundtrip(t *testing.T) {
	src := make([]byte, 2048)
	for i := range src {
		src[i] = byte(i*7 + 3)
	}
	ds := Chunk(src, 256)
	out1, err := ApplyTransforms(ds, CompressLZ4())
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	out2, err := ApplyTransforms(out1, DecompressLZ4())
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if len(out2) != len(ds) {
		t.Fatalf("len mismatch")
	}
	var recon []byte
	for _, d := range out2 {
		recon = append(recon, d.Data...)
	}
	if len(recon) != len(src) {
		t.Fatalf("size mismatch %d != %d", len(recon), len(src))
	}
	for i := range src {
		if recon[i] != src[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}

func TestCompressEncryptDecryptDecompress(t *testing.T) {
	var key [32]byte
	for i := range key {
		key[i] = byte(i + 1)
	}
	a, err := cryptoutil.NewAEAD(key)
	if err != nil {
		t.Fatalf("aead: %v", err)
	}
	enc := &Encryptor{AEAD: a, AAD: []byte("aad")}

	src := make([]byte, 4096)
	for i := range src {
		src[i] = byte(i * 13)
	}
	ds := Chunk(src, 300)
	withSums, err := ApplyTransforms(ds, ComputeChecksum())
	if err != nil {
		t.Fatalf("checksums: %v", err)
	}
	comp, err := ApplyTransforms(withSums, CompressLZ4())
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	encd, err := ApplyTransforms(comp, Encrypt(enc))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decd, err := ApplyTransforms(encd, Decrypt(enc))
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	decomp, err := ApplyTransforms(decd, DecompressLZ4())
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	verified, err := ApplyTransforms(decomp, VerifyChecksum())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	var recon []byte
	for _, d := range verified {
		recon = append(recon, d.Data...)
	}
	if len(recon) != len(src) {
		t.Fatalf("size mismatch")
	}
	for i := range src {
		if recon[i] != src[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}
