package domain_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func TestParseCents(t *testing.T) {
	tests := []struct {
		in      string
		want    unit.Cents
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
		{in: "-", wantErr: true},
		{in: "abc", wantErr: true},
		{in: "22.000", wantErr: true},
		{in: "22.0.0", wantErr: true},
	}

	for _, tc := range tests {
		got, err := unit.ParseCents(tc.in)
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
		in   unit.Cents
		want string
	}{
		{in: 2200, want: "22.00"},
		{in: 471, want: "4.71"},
		{in: 7, want: "0.07"},
		{in: 0, want: "0.00"},
		{in: -471, want: "-4.71"},
	}

	for _, tc := range tests {
		if got := unit.FormatCents(tc.in); got != tc.want {
			t.Errorf("FormatCents(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseGrams(t *testing.T) {
	tests := []struct {
		in      string
		want    unit.Grams
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
		got, err := unit.ParseGrams(tc.in)
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
	got, err := unit.ParseDate("2026-08-01")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if unit.FormatDate(got) != "2026-08-01" {
		t.Errorf("round trip = %q, want 2026-08-01", unit.FormatDate(got))
	}
	for _, in := range []string{"", "01-08-2026", "2026-13-01", "tomorrow"} {
		if _, err := unit.ParseDate(in); err == nil {
			t.Errorf("ParseDate(%q) = nil error, want an error", in)
		}
	}
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		in      string
		want    unit.Percent
		wantErr bool
	}{
		{in: "50", want: 5000},
		{in: "50%", want: 5000},
		{in: " 12.5 % ", want: 1250},
		{in: "12,75", want: 1275},
		{in: "0", want: 0},
		{in: "-1", want: -100},
		{in: "", wantErr: true},
		{in: "-", wantErr: true},
		{in: "half", wantErr: true},
		{in: "12.755", wantErr: true},
	}

	for _, tc := range tests {
		got, err := unit.ParsePercent(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParsePercent(%q) = %d, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePercent(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParsePercent(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		in   unit.Percent
		want string
	}{
		{in: 5000, want: "50%"},
		{in: 1250, want: "12.5%"},
		{in: 1275, want: "12.75%"},
		{in: 0, want: "0%"},
		{in: -100, want: "-1%"},
	}

	for _, tc := range tests {
		if got := unit.FormatPercent(tc.in); got != tc.want {
			t.Errorf("FormatPercent(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseKwhPerHour(t *testing.T) {
	tests := []struct {
		in      string
		want    unit.KwhPerHour
		wantErr bool
	}{
		{in: "0.09", want: 0.09},
		{in: " 0,09 ", want: 0.09},
		{in: "0.09 kWh/h", want: 0.09},
		{in: "0.125", want: 0.125},
		{in: "-0.1", want: -0.1},
		{in: "", wantErr: true},
		{in: "some", wantErr: true},
		{in: "NaN", wantErr: true},
	}

	for _, tc := range tests {
		got, err := unit.ParseKwhPerHour(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseKwhPerHour(%q) = %v, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseKwhPerHour(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseKwhPerHour(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestFormatKwhPerHour(t *testing.T) {
	tests := []struct {
		in   unit.KwhPerHour
		want string
	}{
		{in: 0.09, want: "0.09"},
		{in: 0.1, want: "0.1"},
		{in: 0.125, want: "0.125"},
	}

	for _, tc := range tests {
		if got := unit.FormatKwhPerHour(tc.in); got != tc.want {
			t.Errorf("FormatKwhPerHour(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
