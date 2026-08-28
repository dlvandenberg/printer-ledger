package domain

import (
	"errors"
	"strings"
)

type FilamentType string

const (
	PLA     FilamentType = "PLA"
	PLAPlus FilamentType = "PLA+"
	PETG    FilamentType = "PETG"
)

// ErrMalformedFilamentType names the whole set, so the message cannot drift
// from FilamentTypes. The spool invariant reports the same text.
var ErrMalformedFilamentType = errors.New("must be one of " + strings.Join(FilamentTypeNames(), ", "))

func FilamentTypes() []FilamentType { return []FilamentType{PLA, PLAPlus, PETG} }

func FilamentTypeNames() []string {
	types := FilamentTypes()
	names := make([]string, 0, len(types))
	for _, t := range types {
		names = append(names, t.String())
	}
	return names
}

func ParseFilamentType(s string) (FilamentType, error) {
	t := FilamentType(strings.ToUpper(strings.TrimSpace(s)))
	if !t.Valid() {
		return "", ErrMalformedFilamentType
	}
	return t, nil
}

func (t FilamentType) Valid() bool {
	for _, known := range FilamentTypes() {
		if t == known {
			return true
		}
	}
	return false
}

func (t FilamentType) String() string { return string(t) }
