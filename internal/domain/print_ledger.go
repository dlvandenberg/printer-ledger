package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

var ErrPrintAccountedFor = errors.New("cannot delete a print whose copies are accounted for")

// PrintLedger is a Print with the Sales that reference it counted, so
// availability derives rather than being stored (ADR-0013).
type PrintLedger struct {
	Print     Print
	SoldCount unit.Copies
}

func (l PrintLedger) AccountedCopies() unit.Copies {
	return l.SoldCount + l.Print.GiftedCount + l.Print.KeptCount + l.Print.ScrappedCount
}

func (l PrintLedger) Available() unit.Copies {
	return l.Print.Quantity - l.AccountedCopies()
}

func (l PrintLedger) DeleteBlocked() error {
	if l.AccountedCopies() == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrPrintAccountedFor, l.accountedFor())
}

// validateAvailable is the available >= 0 invariant, which also stops quantity
// shrinking below what has been accounted for (ADR-0013). It reports on
// quantity, the field that can always be raised to satisfy it, and on each
// count that is in the way, so the error also lands on the row the operator
// typed (ADR-0012). It stays quiet where quantity already failed.
func (l PrintLedger) validateAvailable(v *ValidationError) {
	if l.Available() >= 0 {
		return
	}
	if v.For(FieldQuantity) == "" {
		v.Add(FieldQuantity, fmt.Sprintf("cannot be less than the copies accounted for (%s)", l.accountedFor()))
	}
	for _, c := range l.counts() {
		if c.count > 0 && v.For(c.field) == "" {
			v.Add(c.field, "more copies accounted for than this print produced")
		}
	}
}

func (l PrintLedger) accountedFor() string {
	parts := make([]string, 0, len(l.counts())+1)
	if l.SoldCount > 0 {
		parts = append(parts, unit.FormatCopies(l.SoldCount)+" sold")
	}
	for _, c := range l.counts() {
		if c.count > 0 {
			parts = append(parts, unit.FormatCopies(c.count)+" "+c.label)
		}
	}
	return strings.Join(parts, ", ")
}

type recordedCount struct {
	field string
	label string
	count unit.Copies
}

// counts are the outcomes the operator records. Sold is not among them: it
// derives from the Sales referencing this Print and no form owns it.
func (l PrintLedger) counts() []recordedCount {
	return []recordedCount{
		{FieldGifted, "gifted", l.Print.GiftedCount},
		{FieldKept, "kept", l.Print.KeptCount},
		{FieldScrapped, "scrapped", l.Print.ScrappedCount},
	}
}
