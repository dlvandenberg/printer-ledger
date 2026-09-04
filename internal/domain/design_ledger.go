package domain

import (
	"errors"
	"fmt"
)

var ErrDesignPrinted = errors.New("cannot delete a design a print references")

// DesignLedger is a Design with the Prints that reference it counted, so that
// "has this been printed" has one definition rather than one per caller.
type DesignLedger struct {
	Design     Design
	PrintCount int64
}

// DeleteBlocked names the Prints in the way, never the Copies they produced:
// the copies are not what it acts on.
func (l DesignLedger) DeleteBlocked() error {
	if l.PrintCount == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrDesignPrinted, l.printsInTheWay())
}

func (l DesignLedger) printsInTheWay() string {
	if l.PrintCount == 1 {
		return "1 print"
	}
	return fmt.Sprintf("%d prints", l.PrintCount)
}
