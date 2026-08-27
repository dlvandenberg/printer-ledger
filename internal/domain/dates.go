package domain

import (
	"errors"
	"time"
)

const DateLayout = "2006-01-02"

var ErrMalformedDate = errors.New("not a date, expected YYYY-MM-DD")

func ParseDate(s string) (time.Time, error) {
	t, err := time.ParseInLocation(DateLayout, s, time.UTC)
	if err != nil {
		return time.Time{}, ErrMalformedDate
	}
	return t, nil
}

func FormatDate(t time.Time) string { return t.Format(DateLayout) }

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
