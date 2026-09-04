package unit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Grams counts hundredths of a gram — a centigram — so a weight can hold the
// two decimals the slicer reports.
type Grams int64

// GramScale is one gram in hundredths, the factor a Grams has to be divided by
// after it multiplies a rate priced per whole gram.
const GramScale = 100

// maxWholeGrams is the largest weight the gram scale leaves room for, so a
// long digit string is malformed rather than a wrapped negative weight.
const maxWholeGrams = (1<<63 - 1) / GramScale

var ErrMalformedGrams = errors.New("not a whole number of grams")

func ParseGrams(s string) (Grams, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "g")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrMalformedGrams
	}
	g, err := strconv.ParseInt(s, 10, 64)
	if err != nil || g > maxWholeGrams || g < -maxWholeGrams {
		return 0, ErrMalformedGrams
	}
	return Grams(g) * GramScale, nil
}

func FormatGrams(g Grams) string { return fmt.Sprintf("%dg", int64(g)/GramScale) }
