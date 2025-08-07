package pipeline

import (
	"testing"

	"riptide/internal/cryptoutil"
)

func TestEncryptDecryptVerifyPipeline(t *testing.T) {
	var key [32]byte
	for i := range key {
		key[i] = byte(i + 1)
	}
	aead, err := cryptoutil.NewAEAD(key)
	if err != nil {
		t.Fatalf("aead: %v", err)
	}
	enc := &Encryptor{AEAD: aead, AAD: []byte("aad")}
	src := []byte("some test payload data")
	ds := Chunk(src, 5)
	withSums, err := ApplyTransforms(ds, ComputeChecksum())
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	encd, err := ApplyTransforms(withSums, Encrypt(enc))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decd, err := ApplyTransforms(encd, Decrypt(enc))
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	verified, err := ApplyTransforms(decd, VerifyChecksum())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(verified) != len(ds) {
		t.Fatalf("len mismatch")
	}
	var recon []byte
	for _, d := range verified {
		recon = append(recon, d.Data...)
	}
	if string(recon) != string(src) {
		t.Fatalf("roundtrip mismatch: %q != %q", recon, src)
	}
}
