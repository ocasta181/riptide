package congestion

import (
	"testing"
	"time"
)

func TestState_UpdateAndRates(t *testing.T) {
	s := New()
	if s.PacingRate() != 0 {
		t.Fatalf("initial pacing should be 0")
	}
	if cwnd := s.CongestionWindow(1400); cwnd != 1 {
		t.Fatalf("initial cwnd expected 1, got %d", cwnd)
	}

	now := time.Unix(0, 0)
	s.Update(14000, 10*time.Millisecond, 50*time.Millisecond, now)
	if s.PacingRate() <= 0 {
		t.Fatalf("pacing should be > 0")
	}
	if s.minRTT != 50*time.Millisecond {
		t.Fatalf("minRTT not recorded")
	}
	cwnd := s.CongestionWindow(1400)
	if cwnd < 1 {
		t.Fatalf("cwnd should be >= 1, got %d", cwnd)
	}

	// Better bandwidth and lower RTT should increase cwnd
	s.Update(28000, 10*time.Millisecond, 30*time.Millisecond, now.Add(10*time.Millisecond))
	cwnd2 := s.CongestionWindow(1400)
	if cwnd2 < cwnd {
		t.Fatalf("cwnd should not decrease for improved samples: %d -> %d", cwnd, cwnd2)
	}
}

func TestAdjustPayload(t *testing.T) {
	min := 256
	max := 1400

	// Low loss, no corruption: step up
	up := AdjustPayload(500, min, max, 0.0, 0.0)
	if up <= 500 {
		t.Fatalf("expected increase, got %d", up)
	}
	// High loss: step down
	down := AdjustPayload(500, min, max, 0.05, 0.0)
	if down >= 500 {
		t.Fatalf("expected decrease, got %d", down)
	}
	// Corruption: step down
	down2 := AdjustPayload(500, min, max, 0.0, 0.01)
	if down2 >= 500 {
		t.Fatalf("expected decrease on corruption, got %d", down2)
	}
	// Bounds respected
	if got := AdjustPayload(100, min, max, 0.0, 0.0); got < min {
		t.Fatalf("min bound violated: %d", got)
	}
	if got := AdjustPayload(2000, min, max, 0.0, 0.0); got > max {
		t.Fatalf("max bound violated: %d", got)
	}
}
