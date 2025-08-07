package proto

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"riptide/internal/checksum"
)

const (
	Version uint8 = 1
)

type Type uint8

const (
	TypeHello Type = iota + 1
	TypeKX
	TypeAuth
	TypeSession
	TypeData
	TypeAck
	TypeAckAck
	TypeNak
	TypeControl
	TypeFECParity
	TypeHeartbeat
	TypeClose
)

type Header struct {
	Version   uint8
	Type      Type
	Flags     uint16
	Seq       uint64
	Total     uint64
	Timestamp uint64
	Checksum  uint32
}

func (h *Header) Encode() []byte {
	b := make([]byte, 32)
	b[0] = h.Version
	b[1] = byte(h.Type)
	binary.BigEndian.PutUint16(b[2:4], h.Flags)
	binary.BigEndian.PutUint64(b[4:12], h.Seq)
	binary.BigEndian.PutUint64(b[12:20], h.Total)
	binary.BigEndian.PutUint64(b[20:28], h.Timestamp)
	binary.BigEndian.PutUint32(b[28:32], 0)
	cs := crc32.ChecksumIEEE(b[:28])
	binary.BigEndian.PutUint32(b[28:32], cs)
	h.Checksum = cs
	return b
}

func (h *Header) Decode(b []byte) error {
	if len(b) < 32 {
		return errors.New("short header")
	}
	h.Version = b[0]
	h.Type = Type(b[1])
	h.Flags = binary.BigEndian.Uint16(b[2:4])
	h.Seq = binary.BigEndian.Uint64(b[4:12])
	h.Total = binary.BigEndian.Uint64(b[12:20])
	h.Timestamp = binary.BigEndian.Uint64(b[20:28])
	got := binary.BigEndian.Uint32(b[28:32])
	calc := crc32.ChecksumIEEE(b[:28])
	if got != calc {
		return errors.New("bad header checksum")
	}
	h.Checksum = got
	return nil
}

type Ack struct {
	Seq uint64
	Sum checksum.Sum128
}

func (a Ack) Encode() []byte {
	b := make([]byte, 24)
	binary.BigEndian.PutUint64(b[:8], a.Seq)
	copy(b[8:24], a.Sum[:])
	return b
}

func DecodeAck(b []byte) (Ack, error) {
	if len(b) < 24 {
		return Ack{}, errors.New("short ack")
	}
	var a Ack
	a.Seq = binary.BigEndian.Uint64(b[:8])
	copy(a.Sum[:], b[8:24])
	return a, nil
}

type Nak struct {
	Seq  uint64
	Sum  checksum.Sum128
	Code uint16
}

func (n Nak) Encode() []byte {
	b := make([]byte, 26)
	binary.BigEndian.PutUint64(b[:8], n.Seq)
	copy(b[8:24], n.Sum[:])
	binary.BigEndian.PutUint16(b[24:26], n.Code)
	return b
}

func DecodeNak(b []byte) (Nak, error) {
	if len(b) < 26 {
		return Nak{}, errors.New("short nak")
	}
	var n Nak
	n.Seq = binary.BigEndian.Uint64(b[:8])
	copy(n.Sum[:], b[8:24])
	n.Code = binary.BigEndian.Uint16(b[24:26])
	return n, nil
}

type DataPayload struct {
	ChunkID  uint64
	Offset   uint64
	Checksum checksum.Sum128
	Data     []byte
}

func (d DataPayload) Encode() []byte {
	l := 8 + 8 + 16 + len(d.Data)
	b := make([]byte, l)
	binary.BigEndian.PutUint64(b[:8], d.ChunkID)
	binary.BigEndian.PutUint64(b[8:16], d.Offset)
	copy(b[16:32], d.Checksum[:])
	copy(b[32:], d.Data)
	return b
}

func DecodeDataPayload(b []byte) (DataPayload, error) {
	if len(b) < 32 {
		return DataPayload{}, errors.New("short data")
	}
	var d DataPayload
	d.ChunkID = binary.BigEndian.Uint64(b[:8])
	d.Offset = binary.BigEndian.Uint64(b[8:16])
	copy(d.Checksum[:], b[16:32])
	d.Data = make([]byte, len(b)-32)
	copy(d.Data, b[32:])
	return d, nil
}

type HeartbeatPayload struct {
	Seq uint64
}

func (h HeartbeatPayload) Encode() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[:8], h.Seq)
	return b
}

func DecodeHeartbeatPayload(b []byte) (HeartbeatPayload, error) {
	if len(b) < 8 {
		return HeartbeatPayload{}, errors.New("short heartbeat")
	}
	return HeartbeatPayload{Seq: binary.BigEndian.Uint64(b[:8])}, nil
}

type AckAck struct {
	Seq uint64
}

func (a AckAck) Encode() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[:8], a.Seq)
	return b
}

func DecodeAckAck(b []byte) (AckAck, error) {
	if len(b) < 8 {
		return AckAck{}, errors.New("short ack_ack")
	}
	return AckAck{Seq: binary.BigEndian.Uint64(b[:8])}, nil
}

type ControlPayload struct {
	WindowSize uint32
	PacingRate uint32
	RTT        uint64
	LossRate   uint16
	MTUProbe   uint16
}

func (c ControlPayload) Encode() []byte {
	b := make([]byte, 20)
	binary.BigEndian.PutUint32(b[0:4], c.WindowSize)
	binary.BigEndian.PutUint32(b[4:8], c.PacingRate)
	binary.BigEndian.PutUint64(b[8:16], c.RTT)
	binary.BigEndian.PutUint16(b[16:18], c.LossRate)
	binary.BigEndian.PutUint16(b[18:20], c.MTUProbe)
	return b
}

func DecodeControlPayload(b []byte) (ControlPayload, error) {
	if len(b) < 20 {
		return ControlPayload{}, errors.New("short control")
	}
	var c ControlPayload
	c.WindowSize = binary.BigEndian.Uint32(b[0:4])
	c.PacingRate = binary.BigEndian.Uint32(b[4:8])
	c.RTT = binary.BigEndian.Uint64(b[8:16])
	c.LossRate = binary.BigEndian.Uint16(b[16:18])
	c.MTUProbe = binary.BigEndian.Uint16(b[18:20])
	return c, nil
}

type FECParityPayload struct {
	BlockID uint64
	Index   uint16
	Total   uint16
	Parity  []byte
}

func (p FECParityPayload) Encode() []byte {
	b := make([]byte, 12+len(p.Parity))
	binary.BigEndian.PutUint64(b[0:8], p.BlockID)
	binary.BigEndian.PutUint16(b[8:10], p.Index)
	binary.BigEndian.PutUint16(b[10:12], p.Total)
	copy(b[12:], p.Parity)
	return b
}

func DecodeFECParityPayload(b []byte) (FECParityPayload, error) {
	if len(b) < 12 {
		return FECParityPayload{}, errors.New("short fec_parity")
	}
	var p FECParityPayload
	p.BlockID = binary.BigEndian.Uint64(b[0:8])
	p.Index = binary.BigEndian.Uint16(b[8:10])
	p.Total = binary.BigEndian.Uint16(b[10:12])
	p.Parity = make([]byte, len(b)-12)
	copy(p.Parity, b[12:])
	return p, nil
}
