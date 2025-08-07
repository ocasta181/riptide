package netutil

import (
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	IPv4UDPOverhead = 20 + 8
	NonceLen        = 12
)

func UDPPayloadBudget(mtu int) int {
	if mtu <= IPv4UDPOverhead {
		return 0
	}
	return mtu - IPv4UDPOverhead
}

func MaxDataPerPacket(mtu int, headerLen int) int {
	budget := UDPPayloadBudget(mtu)
	overhead := headerLen + NonceLen + chacha20poly1305.Overhead
	if budget <= overhead {
		return 0
	}
	return budget - overhead
}
