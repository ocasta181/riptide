package reliability

import (
	"time"

	"riptide/internal/checksum"
)

type Actions struct {
	ReTx   []uint64
	Ack    []uint64
	AckAck []uint64
}

type ackAckEntry struct {
	seq    uint64
	nextAt time.Time
	tries  int
}

type State struct {
	out        *OutboundTracker
	in         *InboundTracker
	ackAckPend map[uint64]*ackAckEntry
	minAckInt  time.Duration
	lastAck    map[uint64]time.Time
}

func NewState(initialRTO, maxBackoff time.Duration, ackInitialTO, ackMaxBackoff time.Duration, minAckInterval time.Duration, seed int64) *State {
	return &State{
		out:        NewOutboundTracker(initialRTO, maxBackoff, seed),
		in:         NewInboundTracker(ackInitialTO, ackMaxBackoff, seed+1),
		ackAckPend: make(map[uint64]*ackAckEntry),
		minAckInt:  minAckInterval,
		lastAck:    make(map[uint64]time.Time),
	}
}

func (s *State) OnSend(seq uint64, sum checksum.Sum128, now time.Time) {
	s.out.OnSend(seq, sum, now)
}

func (s *State) OnData(seq uint64, sum checksum.Sum128, now time.Time) {
	s.in.OnData(seq, sum, now)
}

func (s *State) OnAck(seq uint64, now time.Time) {
	s.out.OnAck(seq)
	if _, ok := s.ackAckPend[seq]; !ok {
		s.ackAckPend[seq] = &ackAckEntry{seq: seq, nextAt: now}
	}
}

func (s *State) OnAckAck(seq uint64) {
	s.in.OnAckAck(seq)
}

func (s *State) OnNak(seq uint64, now time.Time) {
	s.out.OnNak(seq, now)
}

func (s *State) Tick(now time.Time, max int) Actions {
	var act Actions
	re := s.out.Due(now, max)
	if len(re) > 0 {
		act.ReTx = append(act.ReTx, re...)
	}
	acks := s.in.Due(now, max)
	if len(acks) > 0 {
		for _, seq := range acks {
			if prev, ok := s.lastAck[seq]; ok && now.Sub(prev) < s.minAckInt {
				continue
			}
			act.Ack = append(act.Ack, seq)
			s.lastAck[seq] = now
		}
	}
	if max <= 0 {
		max = len(s.ackAckPend)
	}
	for seq, e := range s.ackAckPend {
		if e.nextAt.After(now) {
			continue
		}
		act.AckAck = append(act.AckAck, seq)
		e.tries++
		to := backoff(s.in.initialTO, e.tries, s.in.maxBackoff)
		j := jitter(s.in.rng, to, 0.1)
		e.nextAt = now.Add(j)
		if len(act.AckAck) == max {
			break
		}
	}
	return act
}

func (s *State) GetOutboundSum(seq uint64) (checksum.Sum128, bool) {
	return s.out.GetSum(seq)
}

func (s *State) GetInboundSum(seq uint64) (checksum.Sum128, bool) {
	return s.in.GetSum(seq)
}
