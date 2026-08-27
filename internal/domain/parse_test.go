package domain_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func TestParseCents(t *testing.T) {
	tests := []struct {
		in      string
		want    domain.Cents
		wantErr bool
	}{
		{in: "22", want: 2200},
		{in: "22.00", want: 2200},
		{in: "22.5", want: 2250},
		{in: "22,50", want: 2250},
		{in: "0.07", want: 7},
		{in: ".5", want: 50},
		{in: "€22.00", want: 2200},
		{in: " 22.00 ", want: 2200},
		{in: "-22.00", want: -2200},
		{in: "", wantErr: true},
		{in: "abc", wantErr: true},
		{in: "22.000", wantErr: true},
		{in: "22.0.0", wantErr: true},
	}

	for _, tc := range tests {
		got, err := domain.ParseCents(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseCents(%q) = %d, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCents(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseCents(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFormatCents(t *testing.T) {
	tests := []struct {
		in   domain.Cents
		want string
	}{
		{in: 2200, want: "22.00"},
		{in: 471, want: "4.71"},
		{in: 7, want: "0.07"},
		{in: 0, want: "0.00"},
		{in: -471, want: "-4.71"},
	}

	for _, tc := range tests {
		if got := domain.FormatCents(tc.in); got != tc.want {
			t.Errorf("FormatCents(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseGrams(t *testing.T) {
	tests := []struct {
		in      string
		want    domain.Grams
		wantErr bool
	}{
		{in: "1000", want: 1000},
		{in: " 210 ", want: 210},
		{in: "750g", want: 750},
		{in: "0", want: 0},
		{in: "-1", want: -1},
		{in: "", wantErr: true},
		{in: "1.5", wantErr: true},
		{in: "lots", wantErr: true},
	}

	for _, tc := range tests {
		got, err := domain.ParseGrams(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseGrams(%q) = %d, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseGrams(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseGrams(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	got, err := domain.ParseDate("2026-08-01")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if domain.FormatDate(got) != "2026-08-01" {
		t.Errorf("round trip = %q, want 2026-08-01", domain.FormatDate(got))
	}
	for _, in := range []string{"", "01-08-2026", "2026-13-01", "tomorrow"} {
		if _, err := domain.ParseDate(in); err == nil {
			t.Errorf("ParseDate(%q) = nil error, want an error", in)
		}
	}
}
