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
	s = strings.TrimPrefix(s, "€")
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
