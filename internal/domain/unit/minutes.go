package unit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const MIN_IN_HOUR = 60

type Minutes int64

var ErrMalformedTime = errors.New("not a time, expected HH:mm")

func ParseTime(s string) (Minutes, error) {
	s = strings.TrimSpace(s)
	hrs, mins, hasColon := strings.Cut(s, ":")
	if !hasColon {
		return 0, ErrMalformedTime
	}

	minutes, err := strconv.Atoi(mins)
	if err != nil {
		return 0, ErrMalformedTime
	}

	if minutes >= MIN_IN_HOUR || minutes < 0 {
		return 0, ErrMalformedTime
	}
	if minutes < 0 {
		return 0, ErrMalformedTime
	}

	hours, err := strconv.Atoi(hrs)
	if err != nil {
		return 0, ErrMalformedTime
	}

	if hours < 0 {
		return 0, ErrMalformedTime
	}

	value := hours*MIN_IN_HOUR + minutes
	return Minutes(value), nil
}

func FormatHHmm(m Minutes) string {
	minutes := m % MIN_IN_HOUR
	hours := (m - minutes) / MIN_IN_HOUR
	return fmt.Sprintf("%d:%02d", hours, minutes)
}

func FormatMinutes(m Minutes) string {
	minutes := m % MIN_IN_HOUR
	hours := (m - minutes) / MIN_IN_HOUR
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
