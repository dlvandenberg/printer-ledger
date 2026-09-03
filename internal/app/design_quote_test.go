package app_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func TestDesignQuotePinsTheReferenceCase(t *testing.T) {
	a := newApp(t)

	if _, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00")); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	row := quoteRow(t, quote, domain.PLA)
	if row.Filament != 132 {
		t.Errorf("Filament = %d, want 132", row.Filament)
	}
	if row.Energy != 7 {
		t.Errorf("Energy = %d, want 7", row.Energy)
	}
	if row.Overhead != 96 {
		t.Errorf("Overhead = %d, want 96", row.Overhead)
	}
	if row.Total != 235 {
		t.Errorf("Total = %d, want 235", row.Total)
	}
	if !quote.HasSuggestedPrice {
		t.Fatal("expected a suggested price")
	}
	if quote.SuggestedPrice != 400 {
		t.Errorf("SuggestedPrice = %d, want 400", quote.SuggestedPrice)
	}
	if quote.SuggestedPriceReference {
		t.Error("expected a live suggested price, not a reference price")
	}
}

func TestDesignQuoteBreaksDownEachFilamentTypeStocked(t *testing.T) {
	a := newApp(t)

	for _, spool := range []app.AddSpoolCmd{
		spoolPriced(domain.PLA, "1000", "22.00"),
		spoolPriced(domain.PETG, "1000", "30.00"),
	} {
		if _, err := a.AddSpool(ctx(), spool); err != nil {
			t.Fatalf("AddSpool: %v", err)
		}
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	if len(quote.Costs) != 2 {
		t.Fatalf("got %d cost rows, want 2", len(quote.Costs))
	}
	if quote.Costs[0].FilamentType != domain.PLA || quote.Costs[1].FilamentType != domain.PETG {
		t.Errorf("rows = %s, %s, want PLA, PETG", quote.Costs[0].FilamentType, quote.Costs[1].FilamentType)
	}

	pla := quoteRow(t, quote, domain.PLA)
	if !pla.IsDefault {
		t.Error("expected the PLA row to be the default filament type")
	}
	petg := quoteRow(t, quote, domain.PETG)
	if petg.IsDefault {
		t.Error("expected the PETG row not to be the default filament type")
	}
	if petg.Filament != 180 {
		t.Errorf("PETG Filament = %d, want 180", petg.Filament)
	}
	if petg.Energy != 9 {
		t.Errorf("PETG Energy = %d, want 9", petg.Energy)
	}
	if petg.Overhead != 96 {
		t.Errorf("PETG Overhead = %d, want 96", petg.Overhead)
	}
	if petg.Total != 285 {
		t.Errorf("PETG Total = %d, want 285", petg.Total)
	}
}

func TestDesignQuoteUsesMostExpensiveCapableSpool(t *testing.T) {
	a := newApp(t)

	cheap, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	expensive, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "30.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	row := quoteRow(t, quote, domain.PLA)
	if row.SpoolID != expensive.ID {
		t.Errorf("SpoolID = %d, want the expensive spool %d", row.SpoolID, expensive.ID)
	}
	if row.Filament != 180 {
		t.Errorf("Filament = %d, want 180", row.Filament)
	}
	if row.Reference {
		t.Error("expected a live price, not a reference price")
	}
	if row.SpoolBrand == "" || row.SpoolColor == "" {
		t.Errorf("expected the row to name its spool, got %q %q", row.SpoolBrand, row.SpoolColor)
	}
	if row.SpoolID == cheap.ID {
		t.Error("expected the cheap spool not to set the price")
	}
}

func TestDesignQuoteExcludesSpoolTooSmallForTheJob(t *testing.T) {
	a := newApp(t)

	cheap, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	expensive, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "30.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if _, err := a.ReweighSpool(ctx(), reweigh(expensive.ID, "240")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	row := quoteRow(t, quote, domain.PLA)
	if row.SpoolID != cheap.ID {
		t.Errorf("SpoolID = %d, want the capable spool %d", row.SpoolID, cheap.ID)
	}
	if row.Filament != 132 {
		t.Errorf("Filament = %d, want 132", row.Filament)
	}
	if row.Reference {
		t.Error("expected a live price, not a reference price")
	}
}

func TestDesignQuoteFallsBackToReferencePriceWithNoCapableSpool(t *testing.T) {
	a := newApp(t)

	cheap, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	expensive, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "30.00"))
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	for _, id := range []int64{cheap.ID, expensive.ID} {
		if _, err := a.ReweighSpool(ctx(), reweigh(id, "240")); err != nil {
			t.Fatalf("ReweighSpool: %v", err)
		}
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	row := quoteRow(t, quote, domain.PLA)
	if !row.Reference {
		t.Error("expected a reference price")
	}
	if row.SpoolID != expensive.ID {
		t.Errorf("SpoolID = %d, want the most expensive spool ever bought %d", row.SpoolID, expensive.ID)
	}
	if row.Filament != 180 {
		t.Errorf("Filament = %d, want 180", row.Filament)
	}
	if quote.SuggestedPrice != 450 {
		t.Errorf("SuggestedPrice = %d, want 450", quote.SuggestedPrice)
	}
	if !quote.SuggestedPriceReference {
		t.Error("expected the suggested price to be flagged a reference price")
	}
}

func TestDesignQuoteOmitsFilamentTypeNeverBought(t *testing.T) {
	a := newApp(t)

	if _, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00")); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	for _, ft := range []domain.FilamentType{domain.PLAPlus, domain.PETG} {
		if hasQuoteRow(quote, ft) {
			t.Errorf("expected no %s row, that type was never bought", ft)
		}
	}
}

func TestDesignQuoteSuggestsPriceFromDefaultFilamentType(t *testing.T) {
	a := newApp(t)

	for _, spool := range []app.AddSpoolCmd{
		spoolPriced(domain.PLA, "1000", "22.00"),
		spoolPriced(domain.PETG, "1000", "30.00"),
	} {
		if _, err := a.AddSpool(ctx(), spool); err != nil {
			t.Fatalf("AddSpool: %v", err)
		}
	}
	cmd := quotedDesign()
	cmd.DefaultFilamentType = domain.PETG.String()
	cmd.MarginPct = "20"
	design, err := a.AddDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	if quote.SuggestedPrice != 350 {
		t.Errorf("SuggestedPrice = %d, want 350", quote.SuggestedPrice)
	}
	if !quoteRow(t, quote, domain.PETG).IsDefault {
		t.Error("expected the PETG row to be the default filament type")
	}
}

func TestDesignQuoteHasNoPriceWhenDefaultTypeNeverBought(t *testing.T) {
	a := newApp(t)

	if _, err := a.AddSpool(ctx(), spoolPriced(domain.PETG, "1000", "30.00")); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	design, err := a.AddDesign(ctx(), quotedDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	if quote.HasSuggestedPrice {
		t.Errorf("expected no suggested price, got %d", quote.SuggestedPrice)
	}
}

func TestDesignQuoteRoundsSuggestedPriceUpToFiftyCents(t *testing.T) {
	tests := []struct {
		name  string
		grams string
		want  unit.Cents
	}{
		{"exact multiple unchanged", "80", 350},
		{"just above a multiple rounds up", "81", 400},
		{"second exact multiple unchanged", "60", 300},
		{"a cent above rounds up", "61", 350},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			if _, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "25.00")); err != nil {
				t.Fatalf("AddSpool: %v", err)
			}
			cmd := quotedDesign()
			cmd.EstimatedGrams = tc.grams
			cmd.EstimatedMinutes = "4:00"
			cmd.MarginPct = "0"
			design, err := a.AddDesign(ctx(), cmd)
			if err != nil {
				t.Fatalf("AddDesign: %v", err)
			}

			quote, err := a.DesignQuote(ctx(), design.ID)
			if err != nil {
				t.Fatalf("DesignQuote: %v", err)
			}

			if quote.SuggestedPrice != tc.want {
				t.Errorf("SuggestedPrice = %d, want %d", quote.SuggestedPrice, tc.want)
			}
		})
	}
}

func TestDesignQuoteRoundsAHalfCentOfEnergyUp(t *testing.T) {
	a := newApp(t)

	settings := settingsUpdate()
	settings.KwhPrice = "1.00"
	settings.PowerRates[domain.PLA] = "0.115"
	if _, err := a.UpdateSettings(ctx(), settings); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if _, err := a.AddSpool(ctx(), spoolPriced(domain.PLA, "1000", "22.00")); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	cmd := quotedDesign()
	cmd.EstimatedMinutes = "1:00"
	design, err := a.AddDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	quote, err := a.DesignQuote(ctx(), design.ID)
	if err != nil {
		t.Fatalf("DesignQuote: %v", err)
	}

	if got := quoteRow(t, quote, domain.PLA).Energy; got != 12 {
		t.Errorf("Energy = %d, want 12", got)
	}
}
