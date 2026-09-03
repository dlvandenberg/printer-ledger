package unit

import (
	"errors"
	"fmt"
	"strings"
)

// Percent is hundredths of a percent, so a margin stays an integer and cost
// math never reaches for a float (ADR-0002).
type Percent int64

// PercentScale is 100% in hundredths of a percent, the factor a Percent has to
// be divided by before it multiplies anything.
const PercentScale = 10_000

var ErrMalformedPercent = errors.New("not a percentage, expected e.g. 50")

func ParsePercent(s string) (Percent, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)

	value, ok := parseHundredths(s)
	if !ok {
		return 0, ErrMalformedPercent
	}
	return Percent(value), nil
}

func FormatPercent(p Percent) string {
	sign := ""
	if p < 0 {
		sign, p = "-", -p
	}
	whole, frac := int64(p)/100, int64(p)%100
	switch {
	case frac == 0:
		return fmt.Sprintf("%s%d%%", sign, whole)
	case frac%10 == 0:
		return fmt.Sprintf("%s%d.%d%%", sign, whole, frac/10)
	default:
		return fmt.Sprintf("%s%d.%02d%%", sign, whole, frac)
	}
}
