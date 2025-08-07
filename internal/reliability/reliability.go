package reliability

import (
	"math/rand"
	"time"

	"riptide/internal/checksum"
)

type OutboundTracker struct {
	entries    map[uint64]*outEntry
	initialRTO time.Duration
	maxBackoff time.Duration
	rng        *rand.Rand
}

type outEntry struct {
	seq    uint64
	sum    checksum.Sum128
	nextAt time.Time
	tries  int
	acked  bool
}

func NewOutboundTracker(initialRTO, maxBackoff time.Duration, seed int64) *OutboundTracker {
	if initialRTO <= 0 {
		initialRTO = 100 * time.Millisecond
	}
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}
	return &OutboundTracker{
		entries:    make(map[uint64]*outEntry),
		initialRTO: initialRTO,
		maxBackoff: maxBackoff,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

func (t *OutboundTracker) OnSend(seq uint64, sum checksum.Sum128, now time.Time) {
	if e, ok := t.entries[seq]; ok {
		e.sum = sum
		if e.acked {
			e.acked = false
		}
		if now.Before(e.nextAt) {
			return
		}
	}
	t.entries[seq] = &outEntry{
		seq:    seq,
		sum:    sum,
		nextAt: now.Add(t.initialRTO),
		tries:  0,
	}
}

func (t *OutboundTracker) OnAck(seq uint64) bool {
	if e, ok := t.entries[seq]; ok {
		e.acked = true
		delete(t.entries, seq)
		return true
	}
	return false
}

func (t *OutboundTracker) OnNak(seq uint64, now time.Time) bool {
	if e, ok := t.entries[seq]; ok {
		e.nextAt = now
		return true
	}
	return false
}

func (t *OutboundTracker) Due(now time.Time, max int) []uint64 {
	if max <= 0 {
		max = len(t.entries)
	}
	out := make([]uint64, 0, max)
	for seq, e := range t.entries {
		if e.nextAt.After(now) {
			continue
		}
		out = append(out, seq)
		e.tries++
		rto := backoff(t.initialRTO, e.tries, t.maxBackoff)
		j := jitter(t.rng, rto, 0.1)
		e.nextAt = now.Add(j)
		if len(out) == cap(out) {
			break
		}
	}
	return out
}

func (t *OutboundTracker) GetSum(seq uint64) (checksum.Sum128, bool) {
	if e, ok := t.entries[seq]; ok {
		return e.sum, true
	}
	return checksum.Sum128{}, false
}

type InboundTracker struct {
	pendingAck map[uint64]*inEntry
	initialTO  time.Duration
	maxBackoff time.Duration
	rng        *rand.Rand
}

type inEntry struct {
	seq    uint64
	sum    checksum.Sum128
	nextAt time.Time
	tries  int
}

func NewInboundTracker(initialTO, maxBackoff time.Duration, seed int64) *InboundTracker {
	if initialTO <= 0 {
		initialTO = 100 * time.Millisecond
	}
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}
	return &InboundTracker{
		pendingAck: make(map[uint64]*inEntry),
		initialTO:  initialTO,
		maxBackoff: maxBackoff,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

func (t *InboundTracker) OnData(seq uint64, sum checksum.Sum128, now time.Time) bool {
	if _, ok := t.pendingAck[seq]; ok {
		return true
	}
	t.pendingAck[seq] = &inEntry{
		seq:    seq,
		sum:    sum,
		nextAt: now.Add(t.initialTO),
		tries:  0,
	}
	return true
}

func (t *InboundTracker) OnAckAck(seq uint64) bool {
	if _, ok := t.pendingAck[seq]; ok {
		delete(t.pendingAck, seq)
		return true
	}
	return false
}

func (t *InboundTracker) Due(now time.Time, max int) []uint64 {
	if max <= 0 {
		max = len(t.pendingAck)
	}
	out := make([]uint64, 0, max)
	for seq, e := range t.pendingAck {
		if e.nextAt.After(now) {
			continue
		}
		out = append(out, seq)
		e.tries++
		to := backoff(t.initialTO, e.tries, t.maxBackoff)
		j := jitter(t.rng, to, 0.1)
		e.nextAt = now.Add(j)
		if len(out) == cap(out) {
			break
		}
	}
	return out
}

func (t *InboundTracker) GetSum(seq uint64) (checksum.Sum128, bool) {
	if e, ok := t.pendingAck[seq]; ok {
		return e.sum, true
	}
	return checksum.Sum128{}, false
}

func backoff(base time.Duration, tries int, max time.Duration) time.Duration {
	d := base << (tries - 1)
	if d > max {
		return max
	}
	return d
}

func jitter(r *rand.Rand, d time.Duration, frac float64) time.Duration {
	if frac <= 0 {
		return d
	}
	n := r.Float64()*2 - 1
	delta := time.Duration(frac * float64(d))
	return d + time.Duration(n*float64(delta))
}
