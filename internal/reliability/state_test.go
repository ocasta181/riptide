package reliability

import (
	"testing"
	"time"

	"riptide/internal/checksum"
)

func TestState_BasicFlows(t *testing.T) {
	now := time.Unix(0, 0)
	st := NewState(
		1*time.Nanosecond, 8*time.Nanosecond,
		1*time.Nanosecond, 8*time.Nanosecond,
		2*time.Nanosecond,
		42,
	)

	s1 := checksum.Compute128([]byte("out"))
	s2 := checksum.Compute128([]byte("in"))

	st.OnSend(1, s1, now)
	st.OnData(10, s2, now)

	acts := st.Tick(now, 10)
	if len(acts.ReTx) != 0 || len(acts.Ack) != 0 || len(acts.AckAck) != 0 {
		t.Fatalf("unexpected actions at t0: %+v", acts)
	}

	now = now.Add(1 * time.Nanosecond)
	acts = st.Tick(now, 10)

	if len(acts.ReTx) != 1 || acts.ReTx[0] != 1 {
		t.Fatalf("expected retransmit of 1, got %+v", acts.ReTx)
	}
	if len(acts.Ack) != 1 || acts.Ack[0] != 10 {
		t.Fatalf("expected ack of 10, got %+v", acts.Ack)
	}
	if _, ok := st.GetOutboundSum(1); !ok {
		t.Fatalf("missing outbound sum")
	}
	if _, ok := st.GetInboundSum(10); !ok {
		t.Fatalf("missing inbound sum")
	}

	st.OnNak(1, now)
	acts = st.Tick(now, 10)
	if len(acts.ReTx) == 0 || acts.ReTx[0] != 1 {
		t.Fatalf("expected retransmit after nak, got %+v", acts.ReTx)
	}

	st.OnAck(1, now)
	acts = st.Tick(now, 10)
	if len(acts.AckAck) == 0 || acts.AckAck[0] != 1 {
		t.Fatalf("expected ack-ack for 1, got %+v", acts.AckAck)
	}

	now = now.Add(1 * time.Nanosecond)
	acts = st.Tick(now, 10)
	if len(acts.Ack) != 0 {
		t.Fatalf("min-interval ack suppression failed: %+v", acts.Ack)
	}
}
