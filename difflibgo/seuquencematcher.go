package difflibgo

import (
	"sort"
)

// sequenceMatcher is a port of the python standard library difflib.SequenceMatcher into go. The
// original class is here: https://github.com/python/cpython/blob/main/Lib/difflib.py#L44. This
// version only works to compare slices of strings, and removes the `junk` components of the python
// implementation.
type sequenceMatcher struct {
	sequenceA      []string
	sequenceB      []string
	matchingBlocks []match
	opCodes        []opCode

	// indices of things in b that are not junk; "b2j" in difflib
	bNonJunkIndicies map[string][]int

	// things deemed "auto junk" by the heuristic; "bpopular" in difflib
	bAutoJunk  map[string]struct{}
	fullBCount map[string]int
}

func (s *sequenceMatcher) setSequences(a, b []string) {
	s.setSequenceA(a)
	s.setSequenceB(b)
}

func (s *sequenceMatcher) setSequenceA(a []string) {
	if &a == &s.sequenceA {
		return
	}

	s.sequenceA = a
	s.matchingBlocks = nil
	s.opCodes = nil
}

func (s *sequenceMatcher) setSequenceB(b []string) {
	if &b == &s.sequenceB {
		return
	}

	s.sequenceB = b
	s.matchingBlocks = nil
	s.opCodes = nil
	s.fullBCount = nil

	s.purgeAutoJunk()
}

func (s *sequenceMatcher) purgeAutoJunkElement() {
	s.bNonJunkIndicies = map[string][]int{}

	seqB := s.sequenceB[0]

	for i, seq := range seqB {
		indices := s.bNonJunkIndicies[string(seq)]
		indices = append(indices, i)
		s.bNonJunkIndicies[string(seq)] = indices
	}

	autoJunk := map[string]struct{}{}

	n := len(seqB)

	if n < autoJunkLenHeuristic {
		return
	}

	ntest := n/oneHundred + 1

	for seq, indices := range s.bNonJunkIndicies {
		if len(indices) > ntest {
			autoJunk[seq] = struct{}{}
		}
	}

	for seq := range autoJunk {
		delete(s.bNonJunkIndicies, seq)
	}

	s.bAutoJunk = autoJunk
}

func (s *sequenceMatcher) purgeAutoJunkSlice() {
	s.bNonJunkIndicies = map[string][]int{}

	for i, seq := range s.sequenceB {
		indices := s.bNonJunkIndicies[seq]
		indices = append(indices, i)
		s.bNonJunkIndicies[seq] = indices
	}

	autoJunk := map[string]struct{}{}

	n := len(s.sequenceB)

	if n < autoJunkLenHeuristic {
		return
	}

	ntest := n/oneHundred + 1

	for seq, indices := range s.bNonJunkIndicies {
		if len(indices) > ntest {
			autoJunk[seq] = struct{}{}
		}
	}

	for seq := range autoJunk {
		delete(s.bNonJunkIndicies, seq)
	}

	s.bAutoJunk = autoJunk
}

func (s *sequenceMatcher) purgeAutoJunk() {
	if len(s.sequenceB) == 1 {
		s.purgeAutoJunkElement()
	} else {
		s.purgeAutoJunkSlice()
	}
}

func (s *sequenceMatcher) isBSeqJunk(seq string) bool {
	_, ok := s.bAutoJunk[seq]

	return ok
}

func (s *sequenceMatcher) findLongestMatchSingleElement(seqALo, seqAHi, seqBLo, seqBHi int) match {
	besti, bestj, bestsize := seqALo, seqBLo, 0
	j2len := map[int]int{}

	seqA, seqB := s.sequenceA[0], s.sequenceB[0]

	if len(seqA) == 0 || len(seqB) == 0 {
		return match{
			A:    0,
			B:    0,
			Size: 0,
		}
	}

	for i := seqALo; i != seqAHi; i++ {
		newj2len := map[int]int{}

		for _, j := range s.bNonJunkIndicies[string(seqA[i])] {
			if j < seqBLo {
				continue
			}

			if j >= seqBHi {
				break
			}

			k := j2len[j-1] + 1
			newj2len[j] = k

			if k > bestsize {
				besti, bestj, bestsize = i-k+1, j-k+1, k
			}
		}

		j2len = newj2len
	}

	for besti > seqALo && bestj > seqBLo && !s.isBSeqJunk(string(seqB[bestj-1])) &&
		string(seqA[besti-1]) == string(seqB[bestj-1]) {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}

	for besti+bestsize < seqAHi && bestj+bestsize < seqBHi &&
		!s.isBSeqJunk(string(seqB[bestj+bestsize])) &&
		string(seqA[besti+bestsize]) == string(seqB[bestj+bestsize]) {
		bestsize++
	}

	for besti > seqALo && bestj > seqBLo && s.isBSeqJunk(string(seqB[bestj-1])) &&
		string(seqA[besti-1]) == string(seqB[bestj-1]) {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}

	for besti+bestsize < seqAHi && bestj+bestsize < seqBHi &&
		s.isBSeqJunk(string(seqB[bestj+bestsize])) &&
		string(seqA[besti+bestsize]) == string(seqB[bestj+bestsize]) {
		bestsize++
	}

	return match{A: besti, B: bestj, Size: bestsize}
}

func (s *sequenceMatcher) findLongestMatchSlice(seqALo, seqAHi, seqBLo, seqBHi int) match {
	besti, bestj, bestsize := seqALo, seqBLo, 0
	j2len := map[int]int{}

	for i := seqALo; i != seqAHi; i++ {
		newj2len := map[int]int{}

		for _, j := range s.bNonJunkIndicies[s.sequenceA[i]] {
			if j < seqBLo {
				continue
			}

			if j >= seqBHi {
				break
			}

			k := j2len[j-1] + 1
			newj2len[j] = k

			if k > bestsize {
				besti, bestj, bestsize = i-k+1, j-k+1, k
			}
		}

		j2len = newj2len
	}

	for besti > seqALo && bestj > seqBLo && !s.isBSeqJunk(s.sequenceB[bestj-1]) &&
		s.sequenceA[besti-1] == s.sequenceB[bestj-1] {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}

	for besti+bestsize < seqAHi && bestj+bestsize < seqBHi &&
		!s.isBSeqJunk(s.sequenceB[bestj+bestsize]) &&
		s.sequenceA[besti+bestsize] == s.sequenceB[bestj+bestsize] {
		bestsize++
	}

	for besti > seqALo && bestj > seqBLo && s.isBSeqJunk(s.sequenceB[bestj-1]) &&
		s.sequenceA[besti-1] == s.sequenceB[bestj-1] {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}

	for besti+bestsize < seqAHi && bestj+bestsize < seqBHi &&
		s.isBSeqJunk(s.sequenceB[bestj+bestsize]) &&
		s.sequenceA[besti+bestsize] == s.sequenceB[bestj+bestsize] {
		bestsize++
	}

	return match{A: besti, B: bestj, Size: bestsize}
}

func (s *sequenceMatcher) findLongestMatch(seqALo, seqAHi, seqBLo, seqBHi int) match {
	if len(s.sequenceA) == 1 && len(s.sequenceB) == 1 {
		return s.findLongestMatchSingleElement(seqALo, seqAHi, seqBLo, seqBHi)
	}

	return s.findLongestMatchSlice(seqALo, seqAHi, seqBLo, seqBHi)
}

func (s *sequenceMatcher) getMatchingBlocks() []match {
	if s.matchingBlocks != nil {
		return s.matchingBlocks
	}

	var matchBlocks func(alo, ahi, blo, bhi int, matched []match) []match

	matchBlocks = func(seqALo, seqAHi, seqBLo, seqBHi int, matched []match) []match {
		longestMatch := s.findLongestMatch(seqALo, seqAHi, seqBLo, seqBHi)
		i, j, k := longestMatch.A, longestMatch.B, longestMatch.Size

		if longestMatch.Size > 0 {
			matched = append(matched, longestMatch)

			if seqALo < i && seqBLo < j {
				matched = matchBlocks(seqALo, i, seqBLo, j, matched)
			}

			if i+k < seqAHi && j+k < seqBHi {
				matched = matchBlocks(i+k, seqAHi, j+k, seqBHi, matched)
			}
		}

		return matched
	}

	la, lb := len(s.sequenceA), len(s.sequenceB)
	if len(s.sequenceA) == 1 && len(s.sequenceB) == 1 {
		la, lb = len(s.sequenceA), len(s.sequenceB)
	}

	matched := matchBlocks(0, la, 0, lb, nil)

	sort.Slice(matched, func(i, j int) bool {
		// this *should*(?) match how the python implementation named tuple sorting works...
		if matched[i].A != matched[j].A {
			return matched[i].A < matched[j].A
		}

		if matched[i].B != matched[j].B {
			return matched[i].B < matched[j].B
		}

		return matched[i].Size < matched[j].Size
	})

	var nonAdjacent []match

	i1, j1, k1 := 0, 0, 0

	for _, b := range matched {
		i2, j2, k2 := b.A, b.B, b.Size
		if i1+k1 == i2 && j1+k1 == j2 {
			k1 += k2
		} else {
			if k1 > 0 {
				nonAdjacent = append(nonAdjacent, match{i1, j1, k1})
			}

			i1, j1, k1 = i2, j2, k2
		}
	}

	if k1 > 0 {
		nonAdjacent = append(nonAdjacent, match{i1, j1, k1})
	}

	nonAdjacent = append(nonAdjacent, match{la, lb, 0})

	s.matchingBlocks = nonAdjacent

	return s.matchingBlocks
}

func (s *sequenceMatcher) getOpcodes() []opCode {
	if s.opCodes != nil {
		return s.opCodes
	}

	i, j := 0, 0
	matching := s.getMatchingBlocks()

	opCodes := make([]opCode, 0, len(matching))

	for _, m := range matching {
		ai, bj, size := m.A, m.B, m.Size
		tag := byte(0)

		switch {
		case i < ai && j < bj:
			tag = 'r'
		case i < ai:
			tag = 'd'
		case j < bj:
			tag = 'i'
		}

		if tag > 0 {
			opCodes = append(opCodes, opCode{tag, i, ai, j, bj})
		}

		i, j = ai+size, bj+size

		if size > 0 {
			opCodes = append(opCodes, opCode{'e', ai, i, bj, j})
		}
	}

	s.opCodes = opCodes

	return s.opCodes
}

func (s *sequenceMatcher) ratio() float64 {
	var la, lb int

	if len(s.sequenceA) == 1 && len(s.sequenceB) == 1 {
		la, lb = len(s.sequenceA[0]), len(s.sequenceB[0])
	} else {
		la, lb = len(s.sequenceA), len(s.sequenceB)
	}

	matches := 0
	for _, mb := range s.getMatchingBlocks() {
		matches += mb.Size
	}

	return calculateRatio(matches, la+lb)
}

func (s *sequenceMatcher) quickRatio() float64 {
	var matches, la, lb int

	if len(s.sequenceA) == 1 && len(s.sequenceB) == 1 { //nolint:nestif
		seqA, seqB := s.sequenceA[0], s.sequenceB[0]
		la, lb = len(seqA), len(seqB)

		if s.fullBCount == nil {
			s.fullBCount = map[string]int{}
			for _, x := range seqB {
				s.fullBCount[string(x)]++
			}
		}

		avail := map[string]int{}
		matches = 0

		for _, x := range seqA {
			n, ok := avail[string(x)]
			if !ok {
				n = s.fullBCount[string(x)]
			}

			avail[string(x)] = n - 1

			if n > 0 {
				matches++
			}
		}
	} else {
		la, lb = len(s.sequenceA), len(s.sequenceB)

		if s.fullBCount == nil {
			s.fullBCount = map[string]int{}
			for _, x := range s.sequenceB {
				s.fullBCount[x]++
			}
		}

		avail := map[string]int{}
		matches = 0
		for _, x := range s.sequenceA {
			n, ok := avail[x]
			if !ok {
				n = s.fullBCount[x]
			}
			avail[x] = n - 1
			if n > 0 {
				matches++
			}
		}
	}

	return calculateRatio(matches, la+lb)
}

func (s *sequenceMatcher) realQuickRatio() float64 {
	var la, lb int

	// different than python because we must have slices of strings, so if slice is len 1
	// we can just use the length of the zeroith element
	if len(s.sequenceA) == 1 && len(s.sequenceB) == 1 {
		la, lb = len(s.sequenceA[0]), len(s.sequenceB[0])
	} else {
		la, lb = len(s.sequenceA), len(s.sequenceB)
	}

	return calculateRatio(min(la, lb), la+lb)
}
