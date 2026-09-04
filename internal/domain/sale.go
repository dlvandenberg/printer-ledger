package domain

import (
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	FieldSalePrint = "printID"
	FieldSalePrice = "price"
	FieldSaleDate  = "date"
)

type Sale struct {
	ID      int64
	PrintID int64
	Price   unit.Cents
	Date    time.Time
}

// NewSale takes the Print Ledger of the Print the copy came from, so the
// available >= 0 invariant is the same expression the Prints list shows
// (ADR-0013). It is what stops the same object being sold twice.
func NewSale(s Sale, l PrintLedger) (Sale, error) {
	v := &ValidationError{}
	validateSale(s, l, v)

	if err := v.OrNil(); err != nil {
		return Sale{}, err
	}
	return s, nil
}

// EditedSale corrects a Sale in place (ADR-0011). The copy the stored Sale
// holds is returned to the Print it came from before availability is checked,
// so re-typing the price of the last copy sold is not an overdraw.
func EditedSale(stored Sale, s Sale, l PrintLedger) (Sale, error) {
	v := &ValidationError{}
	validateSale(s, releaseSale(l, stored), v)
	s.ID = stored.ID

	if err := v.OrNil(); err != nil {
		return Sale{}, err
	}
	return s, nil
}

func validateSale(s Sale, l PrintLedger, v *ValidationError) {
	if s.Price < 0 {
		v.Add(FieldSalePrice, "cannot be negative")
	}
	if s.Date.IsZero() {
		v.Add(FieldSaleDate, "is required")
	}
	switch {
	case s.PrintID == 0:
		v.Add(FieldSalePrint, "is required")
	case l.Available() < 1:
		v.Add(FieldSalePrint, "has no copies available")
	}
}

func releaseSale(l PrintLedger, stored Sale) PrintLedger {
	if stored.PrintID == l.Print.ID {
		l.SoldCount--
	}
	return l
}

// MinimumPrice is the floor a Sale price is measured against: the cost of the
// copy plus the minimum margin.
func MinimumPrice(costPerCopy unit.Cents, minMargin unit.Percent) unit.Cents {
	return withMargin(costPerCopy, minMargin)
}

// BelowFloor warns, it never blocks (ADR-0016). A sale that happened is
// recorded at what was actually charged, haggling included.
func BelowFloor(price, costPerCopy unit.Cents, minMargin unit.Percent) bool {
	return price < MinimumPrice(costPerCopy, minMargin)
}
