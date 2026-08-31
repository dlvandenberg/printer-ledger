package domain

import (
	"strings"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	FieldMeasuredGrams = "measuredGrams"
	FieldAdjustedOn    = "adjustedOn"
	FieldNote          = "note"
)

type SpoolAdjustment struct {
	ID            int64
	SpoolID       int64
	MeasuredGrams unit.Grams
	DeltaGrams    unit.Grams
	AdjustedOn    time.Time
	Note          string
}

func (s SpoolAdjustment) DerivedRemaining(tareGrams unit.Grams) unit.Grams {
	return s.MeasuredGrams - tareGrams
}

func NewSpoolAdjustment(l SpoolLedger, measured unit.Grams, on time.Time, note string) (SpoolAdjustment, error) {
	note = strings.TrimSpace(note)

	v := &ValidationError{}
	if measured < l.Spool.TareGrams {
		v.Add(FieldMeasuredGrams, "cannot be less than the empty spool weight")
	}
	if on.IsZero() {
		v.Add(FieldAdjustedOn, "is required")
	}
	if err := v.OrNil(); err != nil {
		return SpoolAdjustment{}, err
	}

	derived := measured - l.Spool.TareGrams
	return SpoolAdjustment{
		SpoolID:       l.Spool.ID,
		MeasuredGrams: measured,
		DeltaGrams:    derived - l.Remaining(),
		AdjustedOn:    on,
		Note:          note,
	}, nil
}
