package domain

import (
	"strings"
	"time"
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
	InitialGrams Grams
	TareGrams    Grams
	PurchaseCost Cents
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

func (s Spool) ValueOf(g Grams) Cents {
	if s.InitialGrams <= 0 {
		return 0
	}
	return Cents(int64(g) * int64(s.PurchaseCost) / int64(s.InitialGrams))
}
