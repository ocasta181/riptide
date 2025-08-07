package cli

import (
	"reflect"
	"testing"
)

func TestParseArgs_DefaultsAndPositions(t *testing.T) {
	cfg, err := ParseArgs([]string{"src", "dest"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.Src != "src" || cfg.Dest != "dest" {
		t.Fatalf("src/dest mismatch: %+v", cfg)
	}
	if cfg.MTU != 1400 {
		t.Fatalf("default mtu expected 1400, got %d", cfg.MTU)
	}
	wantFEC := FECConfig{Auto: true}
	if !reflect.DeepEqual(cfg.FEC, wantFEC) {
		t.Fatalf("fec mismatch: %+v", cfg.FEC)
	}
	if cfg.Congestion != "bbr" {
		t.Fatalf("default congestion bbr, got %s", cfg.Congestion)
	}
	if cfg.Cipher != "chacha20poly1305" {
		t.Fatalf("default cipher mismatch: %s", cfg.Cipher)
	}
	if cfg.Port != 3703 {
		t.Fatalf("default port 3703, got %d", cfg.Port)
	}
	if cfg.Parallel != 1 || cfg.Resume || cfg.NoCompress || cfg.Checksum || cfg.DryRun {
		t.Fatalf("default flags unexpected: %+v", cfg)
	}
}

func TestParseArgs_AllFlagsValid(t *testing.T) {
	args := []string{
		"-mtu=1200",
		"-fec=4/20",
		"-congestion=ledbat",
		"-id-key=id",
		"-peer-key=peer",
		"-psk=psk",
		"-cipher=chacha20poly1305",
		"-port=4444",
		"-parallel=4",
		"-resume",
		"-no-compress",
		"-checksum",
		"-dry-run",
		"srcX", "destY",
	}
	cfg, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.MTU != 1200 {
		t.Fatalf("mtu mismatch: %d", cfg.MTU)
	}
	if cfg.FEC.Auto || cfg.FEC.K != 4 || cfg.FEC.N != 20 {
		t.Fatalf("fec mismatch: %+v", cfg.FEC)
	}
	if cfg.Congestion != "ledbat" {
		t.Fatalf("congestion mismatch: %s", cfg.Congestion)
	}
	if cfg.IDKey != "id" || cfg.PeerKey != "peer" || cfg.PSK != "psk" {
		t.Fatalf("key fields mismatch: %+v", cfg)
	}
	if cfg.Cipher != "chacha20poly1305" {
		t.Fatalf("cipher mismatch: %s", cfg.Cipher)
	}
	if cfg.Port != 4444 {
		t.Fatalf("port mismatch: %d", cfg.Port)
	}
	if cfg.Parallel != 4 {
		t.Fatalf("parallel mismatch: %d", cfg.Parallel)
	}
	if !cfg.Resume || !cfg.NoCompress || !cfg.Checksum || !cfg.DryRun {
		t.Fatalf("bool flags mismatch: %+v", cfg)
	}
	if cfg.Src != "srcX" || cfg.Dest != "destY" {
		t.Fatalf("src/dest mismatch: %+v", cfg)
	}
}

func TestParseArgs_FEC_AutoAndErrors(t *testing.T) {
	_, err := ParseArgs([]string{"-fec=auto", "a", "b"})
	if err != nil {
		t.Fatalf("auto should be ok: %v", err)
	}
	if _, err := ParseArgs([]string{"-fec=bad", "a", "b"}); err == nil {
		t.Fatalf("expected fec format error")
	}
	if _, err := ParseArgs([]string{"-fec=K/10", "a", "b"}); err == nil {
		t.Fatalf("expected fec K non-numeric error")
	}
	if _, err := ParseArgs([]string{"-fec=4/N", "a", "b"}); err == nil {
		t.Fatalf("expected fec N non-numeric error")
	}
	if _, err := ParseArgs([]string{"-fec=0/10", "a", "b"}); err == nil {
		t.Fatalf("expected invalid fec ratio error")
	}
	if _, err := ParseArgs([]string{"-fec=4/4", "a", "b"}); err == nil {
		t.Fatalf("expected invalid fec ratio K>=N error")
	}
}

func TestParseArgs_Validations(t *testing.T) {
	if _, err := ParseArgs([]string{"-congestion=bad", "a", "b"}); err == nil {
		t.Fatalf("expected congestion error")
	}
	if _, err := ParseArgs([]string{"-cipher=aes", "a", "b"}); err == nil {
		t.Fatalf("expected cipher error")
	}
	if _, err := ParseArgs([]string{"-port=0", "a", "b"}); err == nil {
		t.Fatalf("expected port error")
	}
	if _, err := ParseArgs([]string{"-mtu=0", "a", "b"}); err == nil {
		t.Fatalf("expected mtu error")
	}
	if _, err := ParseArgs([]string{"-parallel=0", "a", "b"}); err == nil {
		t.Fatalf("expected parallel error")
	}
	if _, err := ParseArgs([]string{}); err == nil {
		t.Fatalf("expected positional args error")
	}
}
