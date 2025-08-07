package reliability

import (
	"reflect"
	"testing"
)

func TestBuildSACKRanges_BasicAndDuplicates(t *testing.T) {
	in := []uint64{1, 2, 3, 7, 8, 10, 10, 11}
	out := BuildSACKRanges(in)
	want := []Range{
		{Start: 1, End: 3},
		{Start: 7, End: 8},
		{Start: 10, End: 11},
	}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %+v want %+v", out, want)
	}
}

func TestBuildSACKRanges_UnsortedWithDupes(t *testing.T) {
	in := []uint64{5, 5, 4, 6, 8}
	out := BuildSACKRanges(in)
	want := []Range{
		{Start: 4, End: 6},
		{Start: 8, End: 8},
	}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %+v want %+v", out, want)
	}
}

func TestBuildSACKRanges_Empty(t *testing.T) {
	if out := BuildSACKRanges(nil); out != nil {
		t.Fatalf("expected nil for empty input, got %+v", out)
	}
}
