package unit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Grams int64

var ErrMalformedGrams = errors.New("not a whole number of grams")

func ParseGrams(s string) (Grams, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "g")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrMalformedGrams
	}
	g, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrMalformedGrams
	}
	return Grams(g), nil
}

func FormatGrams(g Grams) string { return fmt.Sprintf("%dg", int64(g)) }
