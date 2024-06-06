package annot

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/rivo/uniseg"
)

// Annot annotates information with a text (represented in Lines)
// at a position in a line (Col).
type Annot struct {
	// Col is the position of the arrowhead of the annotation.
	// E.g. 0 draws an arrow to the first character in a line.
	Col int

	// Lines is the text of the annotation represented in one or more lines.
	Lines []string

	row               int
	lines             []*line
	pipeLeadingSpaces []int
}

type line struct {
	length        int
	leadingSpaces int
}

type section int

// #   |                  column |0123456789...
// ----+-------------------------+-------------------------------------------------
// row | section                 |     ↓ column position of annotation
// -   | not relevant            |     ↑
// 0 · | above · · · · · · · · · | · ██│· · · · · · · · · · · · · · · · · · · · · ·
// 1   | above                   |   ██│
// 2 · | linesFirstTwo · · · · · | · ██└─ line1 · · · · row position of annotation
// 3   | linesFirstTwo           |   █████line2
// 4 · | linesAfterSecond· · · · | ·  · ██line3 · · · · · · · · · · · · · · · · · ·
// 5   | linesAfterSecond        |      ██line4
// 6 · | trailingSpaceLines(1) · | ·  · ███████ · · · · · · · · · · · · · · · · · ·
// 7   | noAnnot                 |
// 8 · | noAnnot · · · · · · · · | ·  · · · · · · · · · · · · · · · · · · · · · · ·
// 9   | noAnnot                 |
// ----+-------------------------+-------------------------------------------------
// #   |                         |    █ = needed space.
const (
	above section = iota
	linesFirstTwo
	linesAfterSecond
	trailingSpaceLines
	noAnnot
)

// AppendLines adds initial or appends additional lines to an annotation.
func (a *Annot) AppendLines(lines ...string) {
	a.Lines = append(a.Lines, lines...)
}

// String returns the rendered annotations as a string.
func String(annots ...*Annot) string {
	b := &strings.Builder{}
	_ = Write(b, annots...)
	return b.String()
}

// Write renders the annotations and writes them to a writer w.
func Write(w io.Writer, annots ...*Annot) error {
	annots = slices.CompactFunc(annots, func(a1 *Annot, a2 *Annot) bool {
		return a1.Col == a2.Col
	})

	if len(annots) == 0 {
		return nil
	}

	slices.SortFunc(annots, func(a *Annot, b *Annot) int {
		return a.Col - b.Col
	})

	for _, a := range annots {
		if len(a.Lines) == 0 {
			a.lines = append(a.lines, &line{})
			continue
		}
		a.lines = make([]*line, len(a.Lines))
		for i := range a.Lines {
			a.lines[i] = &line{
				length:        uniseg.StringWidth(a.Lines[i]),
				leadingSpaces: a.Col,
			}
		}
		a.pipeLeadingSpaces = make([]int, 0)
	}

	// Start with second last annotation index and decrement.
	// The last annotation will always be on row=0 and needs
	// no adjustment.
	for aIdxDecr := len(annots) - 2; 0 <= aIdxDecr; aIdxDecr-- {
		setRow(annots[aIdxDecr], annots[aIdxDecr+1:])
	}

	return write(w, annots)
}

func setRow(a *Annot, rightAnnots []*Annot) {
	row := 0

	for {
		rowBefore := row - 1
		if rowBefore != -1 {
			setSpace(rowBefore, a, rightAnnots)
		}

		annotFits := checkLinesAndSetSpaces(row, a, rightAnnots)
		if annotFits {
			return
		}
		row++
	}
}

func setSpace(rowBefore int, a *Annot, rightAnnots []*Annot) {
	rightA, s := closestAnnot(rowBefore, rightAnnots, 0)
	switch s {
	case above:
		rightA.pipeLeadingSpaces[rowBefore] = rightA.Col - a.Col - 1
	case linesFirstTwo, linesAfterSecond:
		rightA.lines[rowBefore-rightA.row].leadingSpaces = rightA.Col - a.Col - 1
	case noAnnot, trailingSpaceLines:
		// Do nothing
	}
}

func checkLinesAndSetSpaces(row int, a *Annot, rightAnnots []*Annot) bool {
	for aLineIdx := 0; aLineIdx < len(a.lines); aLineIdx++ {
		lineFits := checkLineAndSetSpace(row, aLineIdx, a, rightAnnots)
		if !lineFits {
			return false
		}
	}
	return true
}

func checkLineAndSetSpace(row, aLineIdx int, a *Annot, rightAnnots []*Annot) bool {
	rowPlusLineIdx := row + aLineIdx

	closestA, s := closestAnnot(rowPlusLineIdx, rightAnnots, 1)
	if s == noAnnot {
		return true
	}

	//            3 for "└─ " or "   " (indentation)
	lineLength := 3 + a.lines[aLineIdx].length

	remainingSpaces := closestA.Col - a.Col - lineLength

	//                  |     ↑
	// above            |   ██│
	// above            |   ██│
	// linesFirstTwo    |   ██└─ line1
	// linesFirstTwo    |   ██   line2
	// linesAfterSecond |        line3
	//
	// █ = needed space
	tooLongForAboveOrLinesFirstTwo := (s == above || s == linesFirstTwo) && remainingSpaces < 2

	if remainingSpaces < 0 || tooLongForAboveOrLinesFirstTwo {
		a.row++
		a.pipeLeadingSpaces = append(a.pipeLeadingSpaces, a.Col)
		return false
	}

	switch s {
	case above:
		closestA.pipeLeadingSpaces[rowPlusLineIdx] = remainingSpaces
	case linesFirstTwo, linesAfterSecond:
		closestA.lines[rowPlusLineIdx-closestA.row].leadingSpaces = remainingSpaces
	case trailingSpaceLines:
		closestA2, s2 := closestAnnot(rowPlusLineIdx, rightAnnots, 0)
		if s2 == noAnnot {
			return true
		}
		if s2 == above {
			closestA2.pipeLeadingSpaces[rowPlusLineIdx] = closestA2.Col - a.Col - lineLength
			break
		}
		closestA2.lines[rowPlusLineIdx-closestA2.row].leadingSpaces = closestA2.Col - a.Col - lineLength
	default:
		panic("should never be reached")
	}
	return true
}

func closestAnnot(row int, rightAnnots []*Annot, trailingVerticalSpaceLinesCount int) (*Annot, section) {
	for _, a := range rightAnnots {
		if row < a.row {
			return a, above
		}
		if a.row <= row && row < a.row+len(a.lines) && row < 2 {
			return a, linesFirstTwo
		}
		if 2 <= row && row < a.row+len(a.lines) {
			return a, linesAfterSecond
		}
		if a.row+len(a.lines) <= row && row < a.row+len(a.lines)+trailingVerticalSpaceLinesCount {
			return a, trailingSpaceLines
		}

	}
	return nil, noAnnot
}

func write(writer io.Writer, annots []*Annot) error {
	lastColPos := 0
	rowCount := 0

	b := &strings.Builder{}

	for _, a := range annots {
		b.WriteString(strings.Repeat(" ", a.Col-lastColPos))
		b.WriteString("↑")
		lastColPos = a.Col + 1

		// Also use this loop to calculate rowCount
		rowCount = max(rowCount, a.row+len(a.lines))
	}
	b.WriteString("\n")
	_, err := fmt.Fprint(writer, b.String())
	if err != nil {
		return err
	}
	b.Reset()

	for row := 0; row < rowCount; row++ {
		for _, a := range annots {
			switch {
			case row < a.row:
				b.WriteString(strings.Repeat(" ", a.pipeLeadingSpaces[row]))
				b.WriteString("│")
			case row == a.row:
				b.WriteString(strings.Repeat(" ", a.lines[row-a.row].leadingSpaces))
				b.WriteString("└─ ")
				if len(a.Lines) != 0 {
					b.WriteString(a.Lines[0])
				}
			case row < a.row+len(a.lines):
				b.WriteString(strings.Repeat(" ", a.lines[row-a.row].leadingSpaces))
				b.WriteString("   ")
				b.WriteString(a.Lines[row-a.row])
			}
		}
		b.WriteString("\n")
		_, err = fmt.Fprint(writer, b.String())
		if err != nil {
			return err
		}
		b.Reset()
	}

	return nil
}
