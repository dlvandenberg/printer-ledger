package unit

import (
	"errors"
	"strings"
)

// Grams counts hundredths of a gram — a centigram — so a weight can hold the
// two decimals the slicer reports.
type Grams int64

// GramScale is one gram in hundredths, the factor a Grams has to be divided by
// after it multiplies a rate priced per whole gram.
const GramScale = 100

var ErrMalformedGrams = errors.New("not a weight, expected e.g. 85.59")

func ParseGrams(s string) (Grams, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "g")
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)

	value, ok := parseHundredths(s)
	if !ok {
		return 0, ErrMalformedGrams
	}
	return Grams(value), nil
}

func FormatGrams(g Grams) string { return formatHundredths(int64(g), "g") }
