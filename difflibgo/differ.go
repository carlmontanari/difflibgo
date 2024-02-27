package difflibgo

import (
	"fmt"
	"strings"
	"unicode"
)

// Compare accepts a pair of string slices and returns a slice of their comparison -- this is just
// an itty bitty helper function so you don't need to create a Differ object yourself since it has
// no real value other than its Compare method.
func Compare(seqA, seqB []string) []string {
	d := Differ{}

	return d.Compare(seqA, seqB)
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func calculateRatio(matches, length int) float64 {
	if length > 0 {
		return 2.0 * float64(matches) / float64(length)
	}

	return 1.0
}

type match struct {
	A    int
	B    int
	Size int
}

type opCode struct {
	Tag    byte
	SeqALo int
	SeqAHi int
	SeqBLo int
	SeqBHi int
}

// Differ is an object that helps you compare two string slices.
type Differ struct{}

func (d *Differ) fancyHelper(seqALo, seqAHi, seqBLo, seqBHi int, seqA, seqB []string) []string {
	var g []string

	if seqALo < seqAHi {
		if seqBLo < seqBHi {
			g = d.fancyReplace(seqALo, seqAHi, seqBLo, seqBHi, seqA, seqB)
		} else {
			g = d.dump("-", seqA, seqALo, seqAHi)
		}
	} else if seqBLo < seqBHi {
		g = d.dump("+", seqB, seqBLo, seqBHi)
	}

	return g
}

func (d *Differ) dump(tag string, sequence []string, lo, hi int) []string {
	var dumper []string

	for i := lo; i < hi; i++ {
		dumper = append(dumper, fmt.Sprintf("%s %s", tag, sequence[i]))
	}

	return dumper
}

func keepOriginalWs(s, tags string) string {
	var strippedS string

	var iterLen int

	if len(s) > len(tags) {
		iterLen = len(tags)
	} else {
		iterLen = len(s)
	}

	for i := 0; i < iterLen; i++ {
		c, tagC := s[i], tags[i]

		if string(tagC) == " " && unicode.IsSpace(rune(c)) {
			strippedS += string(c)
		} else {
			strippedS += string(tagC)
		}
	}

	return strings.TrimRight(strippedS, " ")
}

func (d *Differ) qFormat(aline, bline, atags, btags string) []string {
	var f []string

	atags = keepOriginalWs(aline, atags)
	btags = keepOriginalWs(bline, btags)

	f = append(f, fmt.Sprintf("- %s", aline))

	if atags != "" {
		f = append(f, fmt.Sprintf("? %s\n", atags))
	}

	f = append(f, fmt.Sprintf("+ %s", bline))

	if btags != "" {
		f = append(f, fmt.Sprintf("? %s\n", btags))
	}

	return f
}

func (d *Differ) plainReplace(seqALo, seqAHi, seqBLo, seqBHi int, seqA, seqB []string) []string {
	var first []string

	var second []string

	if seqBHi-seqBLo < seqAHi-seqALo {
		first = d.dump("+", seqB, seqBLo, seqBHi)
		second = d.dump("-", seqA, seqALo, seqAHi)
	} else {
		first = d.dump("-", seqA, seqALo, seqAHi)
		second = d.dump("+", seqB, seqBLo, seqBHi)
	}

	return append(first, second...)
}

func assembleFancyReplaceOutput(
	preSyncPointDiffs, formattedTags, postSyncPointDiffs []string,
) []string {
	var finalOut []string

	finalOut = append(finalOut, preSyncPointDiffs...)
	finalOut = append(finalOut, formattedTags...)
	finalOut = append(finalOut, postSyncPointDiffs...)

	return finalOut
}

func (d *Differ) fancyReplace(seqALo, seqAHi, seqBLo, seqBHi int, seqA, seqB []string) []string {
	bestRatio, cutoffRatio := 0.74, 0.75
	eqi, eqj := -1, -1
	bestI, bestJ := -1, -1

	s := &sequenceMatcher{}

	for j := seqBLo; j < seqBHi; j++ {
		bj := seqB[j]

		s.setSequenceB([]string{bj})

		for i := seqALo; i < seqAHi; i++ {
			ai := seqA[i]

			if ai == bj {
				if eqi == -1 {
					eqi, eqj = i, j
				}

				continue
			}

			s.setSequenceA([]string{ai})

			if s.realQuickRatio() > bestRatio && s.quickRatio() > bestRatio &&
				s.ratio() > bestRatio {
				bestRatio = s.ratio()
				bestI, bestJ = i, j
			}
		}
	}

	if bestRatio < cutoffRatio {
		if eqi == -1 {
			replaced := d.plainReplace(seqALo, seqAHi, seqBLo, seqBHi, seqA, seqB)

			return replaced
		}

		bestI, bestJ = eqi, eqj
	} else {
		eqi = -1
	}

	preSyncPointDiffs := d.fancyHelper(seqALo, bestI, seqBLo, bestJ, seqA, seqB)

	aelt, belt := seqA[bestI], seqB[bestJ]

	var formattedTags []string

	if eqi == -1 {
		atags, btags := "", ""

		s.setSequences([]string{aelt}, []string{belt})

		sequenceOpCodes := s.getOpcodes()
		for _, sequenceOpCode := range sequenceOpCodes {
			la := sequenceOpCode.SeqAHi - sequenceOpCode.SeqALo
			lb := sequenceOpCode.SeqBHi - sequenceOpCode.SeqBLo

			switch sequenceOpCode.Tag {
			case replaceOp:
				atags += strings.Repeat("^", la)
				btags += strings.Repeat("^", lb)
			case deleteOp:
				atags += strings.Repeat("-", la)
			case insertOp:
				btags += strings.Repeat("+", lb)
			case equalOp:
				atags += strings.Repeat(" ", la)
				btags += strings.Repeat(" ", lb)
			default:
				panic("unknown opcode, this shouldn't happen...")
			}
		}

		formattedTags = d.qFormat(aelt, belt, atags, btags)
	} else {
		formattedTags = []string{fmt.Sprintf("  %s", aelt)}
	}

	postSyncPointDiffs := d.fancyHelper(seqALo+1, seqAHi, seqBLo+1, seqBHi, seqA, seqB)

	return assembleFancyReplaceOutput(preSyncPointDiffs, formattedTags, postSyncPointDiffs)
}

// Compare accepts two string slices and compares them.
func (d *Differ) Compare(seqA, seqB []string) []string {
	s := &sequenceMatcher{}
	s.setSequences(
		seqA,
		seqB,
	)

	opCodes := s.getOpcodes()

	var finalOut []string

	for _, curOpCode := range opCodes {
		switch curOpCode.Tag {
		case replaceOp:
			c := d.fancyReplace(
				curOpCode.SeqALo,
				curOpCode.SeqAHi,
				curOpCode.SeqBLo,
				curOpCode.SeqBHi,
				seqA,
				seqB,
			)
			finalOut = append(finalOut, c...)
		case deleteOp:
			c := d.dump("-", seqA, curOpCode.SeqALo, curOpCode.SeqAHi)
			finalOut = append(finalOut, c...)
		case insertOp:
			c := d.dump("+", seqB, curOpCode.SeqBLo, curOpCode.SeqBHi)
			finalOut = append(finalOut, c...)
		case equalOp:
			c := d.dump(" ", seqA, curOpCode.SeqALo, curOpCode.SeqAHi)
			finalOut = append(finalOut, c...)
		default:
			panic("unknown opcode, this shouldn't happen...")
		}
	}

	return finalOut
}
