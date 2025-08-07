package queue

import (
	"math/bits"
	"sync/atomic"
)

type Ring[T any] struct {
	_    [64]byte
	cap  uint64
	mask uint64
	buf  []T
	_    [64]byte
	head atomic.Uint64
	_    [64]byte
	tail atomic.Uint64
	_    [64]byte
}

func nextPow2(n int) int {
	if n <= 1 {
		return 1
	}
	if n&(n-1) == 0 {
		return n
	}
	return 1 << bits.Len(uint(n))
}

func NewRing[T any](size int) *Ring[T] {
	if size < 1 {
		size = 1
	}
	c := uint64(nextPow2(size))
	return &Ring[T]{
		cap:  c,
		mask: c - 1,
		buf:  make([]T, int(c)),
	}
}

func (r *Ring[T]) Cap() int {
	return int(r.cap)
}

func (r *Ring[T]) Len() int {
	h := r.head.Load()
	t := r.tail.Load()
	return int(h - t)
}

func (r *Ring[T]) Enqueue(v T) bool {
	h := r.head.Load()
	t := r.tail.Load()
	if h-t >= r.cap {
		return false
	}
	idx := h & r.mask
	r.buf[idx] = v
	r.head.Store(h + 1)
	return true
}

func (r *Ring[T]) Dequeue() (T, bool) {
	var zero T
	t := r.tail.Load()
	h := r.head.Load()
	if t == h {
		return zero, false
	}
	idx := t & r.mask
	v := r.buf[idx]
	var z T
	r.buf[idx] = z
	r.tail.Store(t + 1)
	return v, true
}
