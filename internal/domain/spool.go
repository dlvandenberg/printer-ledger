package domain

import (
	"strings"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	FieldFilamentType = "filamentType"
	FieldBrand        = "brand"
	FieldColor        = "color"
	FieldInitialGrams = "initialGrams"
	FieldTareGrams    = "tareGrams"
	FieldPurchaseCost = "purchaseCost"
	FieldPurchaseDate = "purchaseDate"
)

type SpoolState string

const (
	SpoolActive SpoolState = "active"
	SpoolEmpty  SpoolState = "empty"
)

type Spool struct {
	ID           int64
	FilamentType FilamentType
	Brand        string
	Color        string
	InitialGrams unit.Grams
	TareGrams    unit.Grams
	PurchaseCost unit.Cents
	PurchaseDate time.Time
}

func NewSpool(s Spool) (Spool, error) {
	s.Brand = strings.TrimSpace(s.Brand)
	s.Color = strings.TrimSpace(s.Color)

	v := &ValidationError{}
	if !s.FilamentType.Valid() {
		v.Add(FieldFilamentType, ErrMalformedFilamentType.Error())
	}
	if s.Brand == "" {
		v.Add(FieldBrand, "is required")
	}
	if s.Color == "" {
		v.Add(FieldColor, "is required")
	}
	if s.InitialGrams <= 0 {
		v.Add(FieldInitialGrams, "must be more than 0g")
	}
	if s.TareGrams < 0 {
		v.Add(FieldTareGrams, "cannot be negative")
	}
	if s.PurchaseCost < 0 {
		v.Add(FieldPurchaseCost, "cannot be negative")
	}
	if s.PurchaseDate.IsZero() {
		v.Add(FieldPurchaseDate, "is required")
	}
	if err := v.OrNil(); err != nil {
		return Spool{}, err
	}
	return s, nil
}

func (s Spool) ValueOf(g unit.Grams) unit.Cents {
	if s.InitialGrams <= 0 {
		return 0
	}
	return unit.Cents(int64(g) * int64(s.PurchaseCost) / int64(s.InitialGrams))
}

// QuotedValueOf rounds up where ValueOf truncates, so a quote never sits below
// what the filament actually cost (ADR-0020).
func (s Spool) QuotedValueOf(g unit.Grams) unit.Cents {
	if s.InitialGrams <= 0 {
		return 0
	}
	return unit.Cents(ceilDiv(int64(g)*int64(s.PurchaseCost), int64(s.InitialGrams)))
}

// PricierPerGramThan compares the two ratios by cross-multiplying, so the
// truncation in ValueOf cannot decide which spool sets a quote.
func (s Spool) PricierPerGramThan(o Spool) bool {
	if s.InitialGrams <= 0 || o.InitialGrams <= 0 {
		return s.InitialGrams > 0
	}
	return int64(s.PurchaseCost)*int64(o.InitialGrams) > int64(o.PurchaseCost)*int64(s.InitialGrams)
}
