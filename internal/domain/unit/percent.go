package unit

import (
	"errors"
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

func FormatPercent(p Percent) string { return formatHundredths(int64(p), "%") }
