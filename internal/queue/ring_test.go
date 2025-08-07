package queue

import (
	"sync"
	"testing"
)

func TestRingBasic(t *testing.T) {
	r := NewRing[int](4)
	if r.Cap() < 4 {
		t.Fatalf("cap")
	}
	if _, ok := r.Dequeue(); ok {
		t.Fatalf("expected empty")
	}
	if !r.Enqueue(1) || !r.Enqueue(2) || !r.Enqueue(3) {
		t.Fatalf("enqueue")
	}
	if r.Len() != 3 {
		t.Fatalf("len")
	}
	v, ok := r.Dequeue()
	if !ok || v != 1 {
		t.Fatalf("deq1")
	}
	v, ok = r.Dequeue()
	if !ok || v != 2 {
		t.Fatalf("deq2")
	}
	v, ok = r.Dequeue()
	if !ok || v != 3 {
		t.Fatalf("deq3")
	}
	_, ok = r.Dequeue()
	if ok {
		t.Fatalf("expected empty after deq")
	}
}

func TestRingWraparound(t *testing.T) {
	r := NewRing[int](2)
	if !r.Enqueue(1) || !r.Enqueue(2) {
		t.Fatalf("enqueue fill")
	}
	if r.Enqueue(3) {
		t.Fatalf("should be full")
	}
	v, ok := r.Dequeue()
	if !ok || v != 1 {
		t.Fatalf("deq")
	}
	if !r.Enqueue(3) {
		t.Fatalf("enqueue after wrap")
	}
	if r.Len() != 2 {
		t.Fatalf("len2")
	}
}

func TestRingConcurrentSPSC(t *testing.T) {
	r := NewRing[int](1024)
	var wg sync.WaitGroup
	const N = 10000
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < N; i++ {
			for !r.Enqueue(i) {
			}
		}
	}()
	go func() {
		defer wg.Done()
		exp := 0
		got := 0
		for got < N {
			v, ok := r.Dequeue()
			if !ok {
				continue
			}
			if v != exp {
				t.Fatalf("order %d != %d", v, exp)
			}
			exp++
			got++
		}
	}()
	wg.Wait()
}
