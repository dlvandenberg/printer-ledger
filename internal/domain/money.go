package domain

import (
	"errors"
	"fmt"
	"strings"
)

type Cents int64

var ErrMalformedMoney = errors.New("not an amount, expected e.g. 22.00")

func ParseCents(s string) (Cents, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimLeftFunc(s, isCurrencySymbol)
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)

	value, ok := parseHundredths(s)
	if !ok {
		return 0, ErrMalformedMoney
	}
	return Cents(value), nil
}

func FormatCents(c Cents) string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, int64(c)/100, int64(c)%100)
}

// isCurrencySymbol leads a typed amount with whatever the operator has set as
// the display currency, so €22.00 and $22.00 both read back as 2200.
func isCurrencySymbol(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return false
	case r == '-' || r == '+' || r == '.' || r == ',':
		return false
	default:
		return true
	}
}
