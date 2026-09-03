package domain

import (
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type Design struct {
	ID                  int64
	Name                string
	DefaultFilamentType FilamentType
	EstimatedGrams      unit.Grams
	EstimatedMinutes    unit.Minutes
	MarginPct           unit.Percent
}

const (
	FieldName                = "name"
	FieldEstimatedGrams      = "estimatedGrams"
	FieldEstimatedMinutes    = "estimatedMinutes"
	FieldDefaultFilamentType = "defaultFilamentType"
	FieldMarginPct           = "marginPct"
)

func NewDesign(d Design) (Design, error) {
	d.Name = strings.TrimSpace(d.Name)

	v := &ValidationError{}
	if d.Name == "" {
		v.Add(FieldName, "is required")
	}
	if !d.DefaultFilamentType.Valid() {
		v.Add(FieldDefaultFilamentType, ErrMalformedFilamentType.Error())
	}
	if d.EstimatedGrams <= 0 {
		v.Add(FieldEstimatedGrams, "must be more than 0g")
	}
	if d.EstimatedMinutes <= 0 {
		v.Add(FieldEstimatedMinutes, "must be more than 0min")
	}
	if d.MarginPct < 0 {
		v.Add(FieldMarginPct, "cannot be negative")
	}

	if err := v.OrNil(); err != nil {
		return Design{}, err
	}
	return d, nil
}
