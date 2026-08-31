package unit

import (
	"strconv"
	"strings"
)

// parseHundredths reads a decimal with at most two places into hundredths of a
// unit, so money and percentages accept the same digits and reject the same
// ones.
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
	if err != nil {
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

	value := units*100 + minor
	if negative {
		value = -value
	}
	return value, true
}
