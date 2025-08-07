package proto

import (
	"bytes"
	"testing"

	"riptide/internal/checksum"
)

func TestHeaderEncodeDecode(t *testing.T) {
	h := Header{
		Version:   Version,
		Type:      TypeData,
		Flags:     3,
		Seq:       123,
		Total:     456,
		Timestamp: 789,
	}
	enc := h.Encode()
	var dec Header
	if err := dec.Decode(enc); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if dec != h || dec.Checksum == 0 {
		t.Fatalf("mismatch")
	}
	enc[0] ^= 0xff
	var bad Header
	if err := bad.Decode(enc); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAckEncodeDecode(t *testing.T) {
	s := checksum.Compute128([]byte("x"))
	a := Ack{Seq: 7, Sum: s}
	enc := a.Encode()
	out, err := DecodeAck(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.Seq != a.Seq || !checksum.Equal(out.Sum, a.Sum) {
		t.Fatalf("mismatch")
	}
}

func TestNakEncodeDecode(t *testing.T) {
	s := checksum.Compute128([]byte("y"))
	n := Nak{Seq: 9, Sum: s, Code: 2}
	enc := n.Encode()
	out, err := DecodeNak(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.Seq != n.Seq || out.Code != n.Code || !checksum.Equal(out.Sum, n.Sum) {
		t.Fatalf("mismatch")
	}
}

func TestDataPayloadEncodeDecode(t *testing.T) {
	data := []byte("hello")
	s := checksum.Compute128(data)
	d := DataPayload{
		ChunkID:  11,
		Offset:   22,
		Checksum: s,
		Data:     data,
	}
	enc := d.Encode()
	out, err := DecodeDataPayload(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.ChunkID != d.ChunkID || out.Offset != d.Offset || !bytes.Equal(out.Data, d.Data) || !checksum.Equal(out.Checksum, d.Checksum) {
		t.Fatalf("mismatch")
	}
}

func TestHeartbeatEncodeDecode(t *testing.T) {
	h := HeartbeatPayload{Seq: 99}
	enc := h.Encode()
	out, err := DecodeHeartbeatPayload(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.Seq != h.Seq {
		t.Fatalf("mismatch")
	}
}

func TestAckAckEncodeDecode(t *testing.T) {
	a := AckAck{Seq: 12345}
	enc := a.Encode()
	out, err := DecodeAckAck(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.Seq != a.Seq {
		t.Fatalf("mismatch")
	}
}

func TestControlPayloadEncodeDecode(t *testing.T) {
	c := ControlPayload{
		WindowSize: 1024,
		PacingRate: 2048,
		RTT:        333,
		LossRate:   7,
		MTUProbe:   1400,
	}
	enc := c.Encode()
	out, err := DecodeControlPayload(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out != c {
		t.Fatalf("mismatch")
	}
}

func TestFECParityPayloadEncodeDecode(t *testing.T) {
	parity := []byte{1, 2, 3, 4, 5}
	p := FECParityPayload{
		BlockID: 55,
		Index:   2,
		Total:   8,
		Parity:  parity,
	}
	enc := p.Encode()
	out, err := DecodeFECParityPayload(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.BlockID != p.BlockID || out.Index != p.Index || out.Total != p.Total || !bytes.Equal(out.Parity, p.Parity) {
		t.Fatalf("mismatch")
	}
}
