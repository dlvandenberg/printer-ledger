package app_test

import (
	"path/filepath"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func TestFreshLedgerIsSeededWithSettings(t *testing.T) {
	a := newApp(t)

	got, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if got.KwhPrice != 28 {
		t.Errorf("KwhPrice = %d, want 28", got.KwhPrice)
	}
	if got.MachineHourlyRate != 35 {
		t.Errorf("MachineHourlyRate = %d, want 35", got.MachineHourlyRate)
	}
	if got.DefaultMargin != 5000 {
		t.Errorf("DefaultMargin = %d, want 5000", got.DefaultMargin)
	}
	if got.MinMargin != 1500 {
		t.Errorf("MinMargin = %d, want 1500", got.MinMargin)
	}
	if got.Currency != "€" {
		t.Errorf("Currency = %q, want €", got.Currency)
	}
}

func TestFreshLedgerSeedsOnePowerRatePerFilamentType(t *testing.T) {
	a := newApp(t)

	got, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if len(got.PowerRates) != len(domain.FilamentTypes()) {
		t.Fatalf("got %d power rates, want %d", len(got.PowerRates), len(domain.FilamentTypes()))
	}
	for i, want := range domain.FilamentTypes() {
		rate := got.PowerRates[i]
		if rate.FilamentType != want {
			t.Errorf("power rate %d is for %q, want %q", i, rate.FilamentType, want)
		}
		if rate.KwhPerHour <= 0 {
			t.Errorf("%s rate = %v, want a rate above zero", want, rate.KwhPerHour)
		}
		if rate.Source != domain.RateDefault {
			t.Errorf("%s rate source = %q, want default", want, rate.Source)
		}
	}
}

func TestSettingsAreReadableOnAnEmptyLedger(t *testing.T) {
	a := newApp(t)

	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(spools) != 0 {
		t.Fatalf("got %d spools, want an empty ledger", len(spools))
	}
	if _, err := a.UpdateSettings(ctx(), settingsUpdate()); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
}

func TestUpdateSettingsThenRead(t *testing.T) {
	a := newApp(t)

	cmd := settingsUpdate()
	cmd.KwhPrice = "0.31"
	cmd.MachineHourlyRate = "0.40"
	cmd.PrinterPurchaseCost = "399.00"
	cmd.DefaultMargin = "60"
	cmd.MinMargin = "12.5"
	cmd.Currency = "$"

	updated, err := a.UpdateSettings(ctx(), cmd)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	got, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	for _, s := range []app.SettingsView{updated, got} {
		if s.KwhPrice != 31 {
			t.Errorf("KwhPrice = %d, want 31", s.KwhPrice)
		}
		if s.MachineHourlyRate != 40 {
			t.Errorf("MachineHourlyRate = %d, want 40", s.MachineHourlyRate)
		}
		if s.PrinterPurchaseCost != 39900 {
			t.Errorf("PrinterPurchaseCost = %d, want 39900", s.PrinterPurchaseCost)
		}
		if s.DefaultMargin != 6000 {
			t.Errorf("DefaultMargin = %d, want 6000", s.DefaultMargin)
		}
		if s.MinMargin != 1250 {
			t.Errorf("MinMargin = %d, want 1250", s.MinMargin)
		}
		if s.Currency != "$" {
			t.Errorf("Currency = %q, want $", s.Currency)
		}
	}
}

func TestUpdateSettingsSurvivesAReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")

	first := newAppAt(t, path)
	cmd := settingsUpdate()
	cmd.KwhPrice = "0.31"
	cmd.PowerRates[domain.PETG] = "0.15"
	if _, err := first.UpdateSettings(ctx(), cmd); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	second := newAppAt(t, path)
	got, err := second.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings after reopen: %v", err)
	}
	if got.KwhPrice != 31 {
		t.Errorf("KwhPrice = %d, want 31", got.KwhPrice)
	}
	if rate := powerRate(t, got, domain.PETG); rate.KwhPerHour != 0.15 || rate.Source != domain.RateMeasured {
		t.Errorf("PETG rate = %v %s, want 0.15 measured", rate.KwhPerHour, rate.Source)
	}
}

func TestEditingAPowerRateMarksItMeasured(t *testing.T) {
	a := newApp(t)

	cmd := settingsUpdate()
	cmd.PowerRates[domain.PLA] = "0.11"

	updated, err := a.UpdateSettings(ctx(), cmd)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if rate := powerRate(t, updated, domain.PLA); rate.KwhPerHour != 0.11 || rate.Source != domain.RateMeasured {
		t.Errorf("PLA rate = %v %s, want 0.11 measured", rate.KwhPerHour, rate.Source)
	}
	if rate := powerRate(t, updated, domain.PETG); rate.Source != domain.RateDefault {
		t.Error("PETG rate is measured, want it left as a seeded default")
	}
}

func TestResubmittingAPowerRateUnchangedLeavesItSeeded(t *testing.T) {
	a := newApp(t)

	if _, err := a.UpdateSettings(ctx(), settingsUpdate()); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	got, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	for _, rate := range got.PowerRates {
		if rate.Source != domain.RateDefault {
			t.Errorf("%s rate is measured, want a seeded default", rate.FilamentType)
		}
	}
}

func TestAMeasuredPowerRateStaysMeasured(t *testing.T) {
	a := newApp(t)

	cmd := settingsUpdate()
	cmd.PowerRates[domain.PLA] = "0.11"
	if _, err := a.UpdateSettings(ctx(), cmd); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	again := settingsUpdate()
	again.PowerRates[domain.PLA] = "0.11"
	updated, err := a.UpdateSettings(ctx(), again)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if rate := powerRate(t, updated, domain.PLA); rate.Source != domain.RateMeasured {
		t.Error("PLA rate lost its measured flag when resubmitted unchanged")
	}
}

func TestUpdateSettingsParsesWhatTheOperatorTypes(t *testing.T) {
	a := newApp(t)

	cmd := settingsUpdate()
	cmd.KwhPrice = " €0,31 "
	cmd.DefaultMargin = "60%"
	cmd.PowerRates[domain.PLA] = " 0,11 "

	got, err := a.UpdateSettings(ctx(), cmd)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if got.KwhPrice != 31 {
		t.Errorf("KwhPrice = %d, want 31", got.KwhPrice)
	}
	if got.DefaultMargin != 6000 {
		t.Errorf("DefaultMargin = %d, want 6000", got.DefaultMargin)
	}
	if rate := powerRate(t, got, domain.PLA); rate.KwhPerHour != 0.11 {
		t.Errorf("PLA rate = %v, want 0.11", rate.KwhPerHour)
	}
}

func TestUpdateSettingsValidation(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*app.UpdateSettingsCmd)
		field string
	}{
		{"malformed kwh price", func(c *app.UpdateSettingsCmd) { c.KwhPrice = "cheap" }, domain.FieldKwhPrice},
		{"empty kwh price", func(c *app.UpdateSettingsCmd) { c.KwhPrice = "" }, domain.FieldKwhPrice},
		{"negative kwh price", func(c *app.UpdateSettingsCmd) { c.KwhPrice = "-0.01" }, domain.FieldKwhPrice},
		{"negative machine hourly rate", func(c *app.UpdateSettingsCmd) { c.MachineHourlyRate = "-0.01" }, domain.FieldMachineHourlyRate},
		{"negative printer purchase cost", func(c *app.UpdateSettingsCmd) { c.PrinterPurchaseCost = "-1.00" }, domain.FieldPrinterPurchaseCost},
		{"malformed default margin", func(c *app.UpdateSettingsCmd) { c.DefaultMargin = "half" }, domain.FieldDefaultMargin},
		{"negative default margin", func(c *app.UpdateSettingsCmd) { c.DefaultMargin = "-1" }, domain.FieldDefaultMargin},
		{"negative minimum margin", func(c *app.UpdateSettingsCmd) { c.MinMargin = "-1" }, domain.FieldMinMargin},
		{"empty currency", func(c *app.UpdateSettingsCmd) { c.Currency = "  " }, domain.FieldCurrency},
		{"malformed power rate", func(c *app.UpdateSettingsCmd) { c.PowerRates[domain.PLA] = "a lot" }, domain.FieldPowerRate(domain.PLA)},
		{"zero power rate", func(c *app.UpdateSettingsCmd) { c.PowerRates[domain.PETG] = "0" }, domain.FieldPowerRate(domain.PETG)},
		{"negative power rate", func(c *app.UpdateSettingsCmd) { c.PowerRates[domain.PLAPlus] = "-0.1" }, domain.FieldPowerRate(domain.PLAPlus)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			before, err := a.Settings(ctx())
			if err != nil {
				t.Fatalf("Settings: %v", err)
			}

			cmd := settingsUpdate()
			tc.mut(&cmd)

			if _, err := a.UpdateSettings(ctx(), cmd); err == nil {
				t.Fatal("expected UpdateSettings to be rejected")
			} else if msg := fieldError(t, err, tc.field); msg == "" {
				t.Errorf("no error reported against field %q: %v", tc.field, err)
			}

			after, err := a.Settings(ctx())
			if err != nil {
				t.Fatalf("Settings: %v", err)
			}
			if after.KwhPrice != before.KwhPrice || after.Currency != before.Currency {
				t.Error("a rejected update changed the stored settings")
			}
			for _, rate := range after.PowerRates {
				if rate.Source != domain.RateDefault {
					t.Errorf("a rejected update marked the %s rate measured", rate.FilamentType)
				}
			}
		})
	}
}

func TestUpdateSettingsReportsTheParseFailureNotTheInvariantItTrips(t *testing.T) {
	a := newApp(t)

	cmd := settingsUpdate()
	cmd.KwhPrice = "0.310"

	_, err := a.UpdateSettings(ctx(), cmd)
	if err == nil {
		t.Fatal("expected UpdateSettings to be rejected")
	}
	if msg := fieldError(t, err, domain.FieldKwhPrice); msg != domain.ErrMalformedMoney.Error() {
		t.Errorf("kwhPrice error = %q, want %q", msg, domain.ErrMalformedMoney)
	}
}
