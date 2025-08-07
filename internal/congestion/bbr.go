package congestion

import (
	"math"
	"time"
)

type State struct {
	minRTT       time.Duration
	maxBandwidth float64
	lastTime     time.Time
}

func New() *State {
	return &State{
		minRTT:       time.Hour,
		maxBandwidth: 0,
	}
}

func (s *State) Update(deliveredBytes uint64, interval time.Duration, rttSample time.Duration, now time.Time) {
	if interval > 0 {
		bw := float64(deliveredBytes) / interval.Seconds()
		if bw > s.maxBandwidth {
			s.maxBandwidth = bw
		}
	}
	if rttSample > 0 && rttSample < s.minRTT {
		s.minRTT = rttSample
	}
	s.lastTime = now
}

func (s *State) PacingRate() float64 {
	if s.maxBandwidth <= 0 {
		return 0
	}
	return s.maxBandwidth
}

func (s *State) CongestionWindow(payloadBytes int) int {
	if s.maxBandwidth <= 0 || s.minRTT <= 0 || payloadBytes <= 0 {
		return 1
	}
	bdp := s.maxBandwidth * s.minRTT.Seconds()
	cwnd := int(math.Ceil(bdp / float64(payloadBytes)))
	if cwnd < 1 {
		return 1
	}
	return cwnd
}

func AdjustPayload(current int, min int, max int, lossRate float64, corruptionRate float64) int {
	if current < min {
		current = min
	}
	if current > max {
		current = max
	}
	reduce := lossRate >= 0.02 || corruptionRate > 0
	increase := lossRate < 0.005 && corruptionRate == 0
	stepDown := int(math.Max(1, float64(current)/8))
	stepUp := int(math.Max(1, float64(current)/16))
	if reduce {
		n := current - stepDown
		if n < min {
			n = min
		}
		return n
	}
	if increase {
		n := current + stepUp
		if n > max {
			n = max
		}
		return n
	}
	return current
}
