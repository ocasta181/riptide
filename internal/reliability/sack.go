package reliability

import "sort"

type Range struct {
	Start uint64
	End   uint64
}

func BuildSACKRanges(seqs []uint64) []Range {
	if len(seqs) == 0 {
		return nil
	}
	s := make([]uint64, len(seqs))
	copy(s, seqs)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	out := make([]Range, 0, len(s))
	var curStart, curEnd uint64
	var have bool
	var last uint64
	for i, v := range s {
		if i > 0 && v == last {
			continue
		}
		if !have {
			curStart = v
			curEnd = v
			have = true
		} else if v == curEnd+1 {
			curEnd = v
		} else {
			out = append(out, Range{Start: curStart, End: curEnd})
			curStart = v
			curEnd = v
		}
		last = v
	}
	if have {
		out = append(out, Range{Start: curStart, End: curEnd})
	}
	return out
}
