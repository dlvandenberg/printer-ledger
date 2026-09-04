package domain

import (
	"errors"
	"slices"
	"strings"
	"time"
)

// Period is the span the report is read over. There are no custom ranges: the
// four the operator actually asks about are the whole set.
type Period string

const (
	ThisMonth Period = "this month"
	LastMonth Period = "last month"
	ThisYear  Period = "this year"
	AllTime   Period = "all time"
)

var ErrMalformedPeriod = errors.New("must be one of " + strings.Join(PeriodNames(), ", "))

func Periods() []Period { return []Period{ThisMonth, LastMonth, ThisYear, AllTime} }

func PeriodNames() []string {
	periods := Periods()
	names := make([]string, 0, len(periods))
	for _, p := range periods {
		names = append(names, p.String())
	}
	return names
}

func ParsePeriod(s string) (Period, error) {
	p := Period(strings.ToLower(strings.TrimSpace(s)))
	if !p.Valid() {
		return "", ErrMalformedPeriod
	}
	return p, nil
}

func (p Period) Valid() bool { return slices.Contains(Periods(), p) }

func (p Period) String() string { return string(p) }

// DateRange is half-open: a date on From is in, a date on To is out, so a sale
// on the last day of a month cannot also land in the next one. A zero bound is
// no bound, which is what all time is.
type DateRange struct {
	From time.Time
	To   time.Time
}

func (r DateRange) Bounded() bool { return !r.From.IsZero() && !r.To.IsZero() }

// Through is the last date the range contains, which is the one the operator
// reads: a half-open bound names the day after the period ended.
func (r DateRange) Through() time.Time {
	if r.To.IsZero() {
		return time.Time{}
	}
	return r.To.AddDate(0, 0, -1)
}

func (r DateRange) Contains(t time.Time) bool {
	if !r.From.IsZero() && t.Before(r.From) {
		return false
	}
	return r.To.IsZero() || t.Before(r.To)
}

func (p Period) Range(today time.Time) DateRange {
	month := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	switch p {
	case ThisMonth:
		return DateRange{From: month, To: month.AddDate(0, 1, 0)}
	case LastMonth:
		return DateRange{From: month.AddDate(0, -1, 0), To: month}
	case ThisYear:
		year := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, today.Location())
		return DateRange{From: year, To: year.AddDate(1, 0, 0)}
	case AllTime:
		return DateRange{}
	}
	return DateRange{}
}
