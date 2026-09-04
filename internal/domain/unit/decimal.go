package unit

import (
	"fmt"
	"strconv"
	"strings"
)

// hundredths is the scale every unit in this file shares.
const hundredths = 100

// maxUnits is the largest whole part the hundredths scale leaves room for, two
// decimals included, so a long digit string is malformed rather than a wrapped
// value.
const maxUnits = (1<<63 - 1 - (hundredths - 1)) / hundredths

// parseHundredths reads a decimal with at most two places into hundredths of a
// unit, so money, percentages and weights accept the same digits and reject the
// same ones.
func parseHundredths(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}

	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	if s == "" {
		return 0, false
	}

	whole, frac, hasFrac := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	units, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || units > maxUnits {
		return 0, false
	}

	var minor int64
	if hasFrac {
		switch len(frac) {
		case 1:
			frac += "0"
		case 2:
		default:
			return 0, false
		}
		if minor, err = strconv.ParseInt(frac, 10, 64); err != nil {
			return 0, false
		}
	}

	value := units*hundredths + minor
	if negative {
		value = -value
	}
	return value, true
}

// formatHundredths writes hundredths of a unit with trailing zeros trimmed, so
// every unit on this scale displays the same way.
func formatHundredths(value int64, suffix string) string {
	sign := ""
	if value < 0 {
		sign, value = "-", -value
	}
	whole, frac := value/hundredths, value%hundredths
	switch {
	case frac == 0:
		return fmt.Sprintf("%s%d%s", sign, whole, suffix)
	case frac%10 == 0:
		return fmt.Sprintf("%s%d.%d%s", sign, whole, frac/10, suffix)
	default:
		return fmt.Sprintf("%s%d.%02d%s", sign, whole, frac, suffix)
	}
}
