package reliability

import (
	"testing"
	"time"

	"riptide/internal/checksum"
)

func TestOutboundTracker_BasicFlow(t *testing.T) {
	sum := checksum.Compute128([]byte("a"))
	tr := NewOutboundTracker(1*time.Nanosecond, 16*time.Nanosecond, 1)
	now := time.Unix(0, 0)

	tr.OnSend(1, sum, now)
	got, ok := tr.GetSum(1)
	if !ok || !checksum.Equal(got, sum) {
		t.Fatalf("missing or wrong sum")
	}
	if due := tr.Due(now, 10); len(due) != 0 {
		t.Fatalf("should not be due yet")
	}

	now = now.Add(1 * time.Nanosecond)
	due := tr.Due(now, 10)
	if len(due) != 1 || due[0] != 1 {
		t.Fatalf("expected seq 1 due, got %v", due)
	}

	now = now.Add(1 * time.Nanosecond)
	due = tr.Due(now, 10)
	if len(due) != 1 || due[0] != 1 {
		t.Fatalf("expected seq 1 due again, got %v", due)
	}

	if !tr.OnAck(1) {
		t.Fatalf("ack should remove entry")
	}
	now = now.Add(10 * time.Nanosecond)
	if due := tr.Due(now, 10); len(due) != 0 {
		t.Fatalf("no entries should be due after ack")
	}
	if _, ok := tr.GetSum(1); ok {
		t.Fatalf("sum should be gone after ack")
	}
}

func TestOutboundTracker_NakImmediate(t *testing.T) {
	sum := checksum.Compute128([]byte("b"))
	tr := NewOutboundTracker(1*time.Nanosecond, 16*time.Nanosecond, 2)
	now := time.Unix(0, 0)

	tr.OnSend(2, sum, now)
	now1 := now.Add(1 * time.Nanosecond)
	due := tr.Due(now1, 10)
	if len(due) != 1 || due[0] != 2 {
		t.Fatalf("expected seq 2 due first time, got %v", due)
	}

	if !tr.OnNak(2, now1) {
		t.Fatalf("nak should be applied")
	}
	due = tr.Due(now1, 10)
	if len(due) != 1 || due[0] != 2 {
		t.Fatalf("expected immediate re-due after nak, got %v", due)
	}
}

func TestInboundTracker_BasicFlow(t *testing.T) {
	sum := checksum.Compute128([]byte("c"))
	tr := NewInboundTracker(1*time.Nanosecond, 16*time.Nanosecond, 3)
	now := time.Unix(0, 0)

	if !tr.OnData(5, sum, now) {
		t.Fatalf("on data failed")
	}
	got, ok := tr.GetSum(5)
	if !ok || !checksum.Equal(got, sum) {
		t.Fatalf("missing or wrong sum")
	}
	if due := tr.Due(now, 10); len(due) != 0 {
		t.Fatalf("should not be due yet")
	}

	now = now.Add(1 * time.Nanosecond)
	due := tr.Due(now, 10)
	if len(due) != 1 || due[0] != 5 {
		t.Fatalf("expected seq 5 due, got %v", due)
	}

	if !tr.OnAckAck(5) {
		t.Fatalf("ack-ack should remove pending entry")
	}
	now = now.Add(10 * time.Nanosecond)
	if due := tr.Due(now, 10); len(due) != 0 {
		t.Fatalf("no entries should be due after ack-ack")
	}
	if _, ok := tr.GetSum(5); ok {
		t.Fatalf("sum should be gone after ack-ack")
	}
}
