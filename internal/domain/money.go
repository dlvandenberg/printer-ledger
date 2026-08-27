package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Cents int64

var ErrMalformedMoney = errors.New("not an amount, expected e.g. 22.00")

func ParseCents(s string) (Cents, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "€")
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)
	if s == "" {
		return 0, ErrMalformedMoney
	}

	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	whole, frac, hasFrac := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	units, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, ErrMalformedMoney
	}

	var minor int64
	if hasFrac {
		switch len(frac) {
		case 1:
			frac += "0"
		case 2:
		default:
			return 0, ErrMalformedMoney
		}
		if minor, err = strconv.ParseInt(frac, 10, 64); err != nil {
			return 0, ErrMalformedMoney
		}
	}

	cents := Cents(units*100 + minor)
	if negative {
		cents = -cents
	}
	return cents, nil
}

func FormatCents(c Cents) string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, int64(c)/100, int64(c)%100)
}
