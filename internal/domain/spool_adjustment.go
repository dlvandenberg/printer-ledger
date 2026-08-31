package domain

import (
	"strings"
	"time"
)

const (
	FieldMeasuredGrams = "measuredGrams"
	FieldAdjustedOn    = "adjustedOn"
	FieldNote          = "note"
)

type SpoolAdjustment struct {
	ID            int64
	SpoolID       int64
	MeasuredGrams Grams
	DeltaGrams    Grams
	AdjustedOn    time.Time
	Note          string
}

func (s SpoolAdjustment) DerivedRemaining(tareGrams Grams) Grams {
	return s.MeasuredGrams - tareGrams
}

func NewSpoolAdjustment(l SpoolLedger, measured Grams, on time.Time, note string) (SpoolAdjustment, error) {
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
