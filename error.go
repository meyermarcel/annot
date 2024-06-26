package annot

import (
	"errors"
	"fmt"
)

type OverlapError struct {
	colEnd, firstAnnotPos, secondAnnotCol int
}

func newOverlapError(colEnd, firstAnnotPos, secondAnnotCol int) *OverlapError {
	return &OverlapError{colEnd, firstAnnotPos, secondAnnotCol}
}

func (e *OverlapError) Error() string {
	return fmt.Sprintf("annot: ColEnd %d of %d. annotation overlaps with Col %d of %d. annotation",
		e.colEnd, e.firstAnnotPos, e.secondAnnotCol, e.firstAnnotPos+1)
}

func (e *OverlapError) Is(target error) bool {
	var overlapError *OverlapError
	return errors.As(target, &overlapError)
}

type ColExceedsColEndError struct {
	annotPos, col, colEnd int
}

func newColExceedsColEndError(annotPos, col, colEnd int) *ColExceedsColEndError {
	return &ColExceedsColEndError{annotPos, col, colEnd}
}

func (e *ColExceedsColEndError) Error() string {
	return fmt.Sprintf("annot: in %d. annotation Col %d needs to be lower than ColEnd %d",
		e.annotPos, e.col, e.colEnd)
}

func (e *ColExceedsColEndError) Is(target error) bool {
	var overlapError *ColExceedsColEndError
	return errors.As(target, &overlapError)
}
