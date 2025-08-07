package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type FECConfig struct {
	Auto bool
	K    int
	N    int
}

type Config struct {
	Src        string
	Dest       string
	MTU        int
	FEC        FECConfig
	Congestion string
	IDKey      string
	PeerKey    string
	PSK        string
	Cipher     string
	Port       int
	Parallel   int
	Resume     bool
	NoCompress bool
	Checksum   bool
	DryRun     bool
}

func ParseArgs(args []string) (Config, error) {
	var cfg Config
	fs := flag.NewFlagSet("riptide", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.IntVar(&cfg.MTU, "mtu", 1400, "payload sizing ceiling")
	fecStr := fs.String("fec", "auto", "fec ratio k/n or 'auto'")
	fs.StringVar(&cfg.Congestion, "congestion", "bbr", "congestion controller")
	fs.StringVar(&cfg.IDKey, "id-key", "", "identity key path")
	fs.StringVar(&cfg.PeerKey, "peer-key", "", "peer public key")
	fs.StringVar(&cfg.PSK, "psk", "", "pre-shared key path")
	fs.StringVar(&cfg.Cipher, "cipher", "chacha20poly1305", "cipher")
	fs.IntVar(&cfg.Port, "port", 3703, "udp port")
	fs.IntVar(&cfg.Parallel, "parallel", 1, "parallelism factor")
	fs.BoolVar(&cfg.Resume, "resume", false, "resume transfers")
	fs.BoolVar(&cfg.NoCompress, "no-compress", false, "disable compression")
	fs.BoolVar(&cfg.Checksum, "checksum", false, "force strong checksum compare")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "plan only")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	rest := fs.Args()
	if len(rest) != 2 {
		return Config{}, errors.New("expected SRC and DEST")
	}
	cfg.Src = rest[0]
	cfg.Dest = rest[1]

	fec, err := parseFEC(*fecStr)
	if err != nil {
		return Config{}, err
	}
	cfg.FEC = fec

	if err := validate(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validate(c *Config) error {
	if c.MTU <= 0 {
		return errors.New("mtu must be > 0")
	}
	switch c.Congestion {
	case "bbr", "ledbat":
	default:
		return fmt.Errorf("invalid congestion: %s", c.Congestion)
	}
	if c.Cipher != "chacha20poly1305" {
		return fmt.Errorf("invalid cipher: %s", c.Cipher)
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	if c.Parallel <= 0 {
		return errors.New("parallel must be > 0")
	}
	if !c.FEC.Auto {
		if c.FEC.K <= 0 || c.FEC.N <= 0 || c.FEC.K >= c.FEC.N {
			return errors.New("invalid fec ratio")
		}
	}
	return nil
}

func parseFEC(s string) (FECConfig, error) {
	if s == "auto" {
		return FECConfig{Auto: true}, nil
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return FECConfig{}, errors.New("fec must be k/n or auto")
	}
	k, err := strconv.Atoi(parts[0])
	if err != nil {
		return FECConfig{}, errors.New("fec k invalid")
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return FECConfig{}, errors.New("fec n invalid")
	}
	return FECConfig{K: k, N: n}, nil
}
