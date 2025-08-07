package pipeline

import (
	"reflect"
	"testing"

	"riptide/internal/checksum"
	"riptide/internal/queue"
)

func TestComposeAndTransforms(t *testing.T) {
	d := Descriptor{ChunkID: 1, Offset: 2, Data: []byte("abc")}
	t1 := func(in Descriptor) (Descriptor, error) {
		in.Sum = checksum.Compute128(in.Data)
		return in, nil
	}
	t2 := func(in Descriptor) (Descriptor, error) {
		out := in
		tmp := make([]byte, len(in.Data)+1)
		copy(tmp, in.Data)
		tmp[len(in.Data)] = 'x'
		out.Data = tmp
		out.Sum = checksum.Compute128(out.Data)
		return out, nil
	}
	tr := Compose(t1, t2)
	out, err := tr(d)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(out.Data) != "abcx" {
		t.Fatalf("data mismatch: %q", out.Data)
	}
	want := checksum.Compute128([]byte("abcx"))
	if !checksum.Equal(out.Sum, want) {
		t.Fatalf("sum mismatch")
	}
	if !reflect.DeepEqual(d.Data, []byte("abc")) {
		t.Fatalf("input mutated")
	}
}

func TestComposeNilTransform(t *testing.T) {
	tr := Compose(nil)
	_, err := tr(Descriptor{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestQueueInterfaceCompatibility(t *testing.T) {
	r := queue.NewRing[int](8)
	var q Queue[int] = r
	if !q.Enqueue(3) {
		t.Fatalf("enqueue failed")
	}
	v, ok := q.Dequeue()
	if !ok || v != 3 {
		t.Fatalf("dequeue mismatch")
	}
}
