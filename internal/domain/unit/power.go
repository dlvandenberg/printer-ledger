package unit

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// KwhPerHour is a measured physical rate, not money, so it stays a float64.
type KwhPerHour float64

var ErrMalformedKwhPerHour = errors.New("not a power rate, expected e.g. 0.09")

func ParseKwhPerHour(s string) (KwhPerHour, error) {
	s = strings.TrimSpace(s)
	if len(s) >= 5 && strings.EqualFold(s[len(s)-5:], "kwh/h") {
		s = s[:len(s)-5]
	}
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)
	if s == "" {
		return 0, ErrMalformedKwhPerHour
	}

	rate, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(rate) || math.IsInf(rate, 0) {
		return 0, ErrMalformedKwhPerHour
	}
	return KwhPerHour(rate), nil
}

func FormatKwhPerHour(k KwhPerHour) string {
	return strconv.FormatFloat(float64(k), 'f', -1, 64)
}
