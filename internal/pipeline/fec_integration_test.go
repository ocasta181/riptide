package pipeline

import "testing"

func TestFECGroupEncodeReconstruct(t *testing.T) {
	src := make([]byte, 1024)
	for i := range src {
		src[i] = byte(255 - (i % 256))
	}
	ds := Chunk(src, 256)
	if len(ds) < 4 {
		t.Fatalf("need at least 4 chunks")
	}
	group := ds[:4]

	enc, err := FECGroupEncode(group, 4, 2)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(enc) != 6 {
		t.Fatalf("expected 6 shards, got %d", len(enc))
	}
	// Lose one data shard (index 2)
	lost := 2
	mut := make([]Descriptor, len(enc))
	copy(mut, enc)
	mut[lost].Data = nil

	recon, err := FECGroupReconstruct(mut, 4, 2, []int{lost})
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if len(recon) != len(enc) {
		t.Fatalf("len mismatch")
	}
	for i := 0; i < 4; i++ {
		if len(recon[i].Data) != len(enc[i].Data) {
			t.Fatalf("shard %d size mismatch", i)
		}
		for j := range recon[i].Data {
			if recon[i].Data[j] != enc[i].Data[j] {
				t.Fatalf("shard %d byte mismatch at %d", i, j)
			}
		}
	}
}
