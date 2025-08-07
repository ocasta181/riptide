package proto

import (
	"errors"

	"riptide/internal/cryptoutil"
)

const headerLen = 32
const nonceLen = 12

func EncodeDataPacket(h Header, payload DataPayload, a *cryptoutil.AEAD, aad []byte) ([]byte, error) {
	hb := h.Encode()
	pb := payload.Encode()
	ct, nonce := a.Seal(nil, pb, aad)
	out := make([]byte, 0, len(hb)+nonceLen+len(ct))
	out = append(out, hb...)
	out = append(out, nonce[:]...)
	out = append(out, ct...)
	return out, nil
}

func DecodeDataPacket(b []byte, a *cryptoutil.AEAD, aad []byte) (Header, DataPayload, error) {
	if len(b) < headerLen+nonceLen {
		return Header{}, DataPayload{}, errors.New("short packet")
	}
	var h Header
	if err := h.Decode(b[:headerLen]); err != nil {
		return Header{}, DataPayload{}, err
	}
	var n [nonceLen]byte
	copy(n[:], b[headerLen:headerLen+nonceLen])
	ct := b[headerLen+nonceLen:]
	pt, err := a.Open(nil, ct, aad, n)
	if err != nil {
		return Header{}, DataPayload{}, err
	}
	dp, err := DecodeDataPayload(pt)
	if err != nil {
		return Header{}, DataPayload{}, err
	}
	return h, dp, nil
}
