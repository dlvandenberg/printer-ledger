package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func TestAddDesignThenList(t *testing.T) {
	a := newApp(t)

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}
	if added.ID == 0 {
		t.Error("expected the added design to have an ID")
	}

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 1 {
		t.Fatalf("got %d designs, want 1", len(designs))
	}

	got := designs[0]
	if got.ID != added.ID {
		t.Errorf("ID = %d, want %d", got.ID, added.ID)
	}
	if got.Name != "Cable clip" {
		t.Errorf("Name = %q, want Cable clip", got.Name)
	}
	if got.DefaultFilamentType != domain.PLA {
		t.Errorf("DefaultFilamentType = %q, want PLA", got.DefaultFilamentType)
	}
	if got.EstimatedGrams != grams(48) {
		t.Errorf("EstimatedGrams = %d, want %d", got.EstimatedGrams, grams(48))
	}
	if got.EstimatedMinutes != 331 {
		t.Errorf("EstimatedMinutes = %d, want 331", got.EstimatedMinutes)
	}
}

func TestAddDesignParsesPrintTime(t *testing.T) {
	tests := []struct {
		name  string
		typed string
		want  unit.Minutes
	}{
		{"hours and minutes", "5:31", 331},
		{"zero hours", "0:45", 45},
		{"padded", "  12:00  ", 720},
		{"over a day", "26:05", 1565},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			cmd := plaDesign()
			cmd.EstimatedMinutes = tc.typed

			added, err := a.AddDesign(ctx(), cmd)
			if err != nil {
				t.Fatalf("AddDesign: %v", err)
			}
			if added.EstimatedMinutes != tc.want {
				t.Errorf("EstimatedMinutes = %d, want %d", added.EstimatedMinutes, tc.want)
			}
		})
	}
}

func TestAddDesignSeedsMarginFromSettings(t *testing.T) {
	a := newApp(t)

	settings, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}

	cmd := plaDesign()
	cmd.MarginPct = ""

	added, err := a.AddDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}
	if added.MarginPct != settings.DefaultMargin {
		t.Errorf("MarginPct = %d, want the settings default %d", added.MarginPct, settings.DefaultMargin)
	}
}

func TestAddDesignSeedsTheDefaultAsItStandsNow(t *testing.T) {
	a := newApp(t)

	raised := settingsUpdate()
	raised.DefaultMargin = "70"
	if _, err := a.UpdateSettings(ctx(), raised); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}
	if added.MarginPct != 7000 {
		t.Errorf("MarginPct = %d, want 7000", added.MarginPct)
	}
}

func TestAddDesignKeepsATypedMargin(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.MarginPct = "65%"

	added, err := a.AddDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}
	if added.MarginPct != 6500 {
		t.Errorf("MarginPct = %d, want 6500", added.MarginPct)
	}
}

func TestRaisingTheDefaultMarginLeavesExistingDesignsAlone(t *testing.T) {
	a := newApp(t)

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	raised := settingsUpdate()
	raised.DefaultMargin = "80"
	if _, err := a.UpdateSettings(ctx(), raised); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	got := designNamed(t, designs, "Cable clip")
	if got.MarginPct != added.MarginPct {
		t.Errorf("MarginPct = %d, want it unchanged at %d", got.MarginPct, added.MarginPct)
	}
}

func TestEditDesignUpdatesEstimatesAfterAReslice(t *testing.T) {
	a := newApp(t)

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	cmd := editOf(added)
	cmd.EstimatedGrams = "52"
	cmd.EstimatedMinutes = "6:02"
	cmd.DefaultFilamentType = domain.PETG.String()
	cmd.Name = "Cable clip v2"

	if _, err := a.EditDesign(ctx(), cmd); err != nil {
		t.Fatalf("EditDesign: %v", err)
	}

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 1 {
		t.Fatalf("got %d designs, want 1", len(designs))
	}

	got := designs[0]
	if got.Name != "Cable clip v2" {
		t.Errorf("Name = %q, want Cable clip v2", got.Name)
	}
	if got.EstimatedGrams != grams(52) {
		t.Errorf("EstimatedGrams = %d, want %d", got.EstimatedGrams, grams(52))
	}
	if got.EstimatedMinutes != 362 {
		t.Errorf("EstimatedMinutes = %d, want 362", got.EstimatedMinutes)
	}
	if got.DefaultFilamentType != domain.PETG {
		t.Errorf("DefaultFilamentType = %q, want PETG", got.DefaultFilamentType)
	}
	if got.MarginPct != added.MarginPct {
		t.Errorf("MarginPct = %d, want it unchanged at %d", got.MarginPct, added.MarginPct)
	}
}

func TestEditDesignChangesTheMarginOnItsOwn(t *testing.T) {
	a := newApp(t)

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	cmd := editOf(added)
	cmd.MarginPct = "35"

	edited, err := a.EditDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("EditDesign: %v", err)
	}
	if edited.MarginPct != 3500 {
		t.Errorf("MarginPct = %d, want 3500", edited.MarginPct)
	}
	if edited.EstimatedGrams != added.EstimatedGrams {
		t.Errorf("EstimatedGrams = %d, want it unchanged at %d", edited.EstimatedGrams, added.EstimatedGrams)
	}

	settings, err := a.Settings(ctx())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if settings.DefaultMargin != domain.DefaultSettings().DefaultMargin {
		t.Errorf("DefaultMargin = %d, want editing a design to leave settings alone", settings.DefaultMargin)
	}
}

func TestEditDesignDoesNotReseedABlankMargin(t *testing.T) {
	a := newApp(t)

	added, err := a.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	cmd := editOf(added)
	cmd.MarginPct = ""

	if _, err := a.EditDesign(ctx(), cmd); err == nil {
		t.Fatal("expected EditDesign to be rejected")
	} else if msg := fieldError(t, err, domain.FieldMarginPct); msg == "" {
		t.Errorf("no error reported against the margin: %v", err)
	}
}

func TestEditDesignRejectsAnUnknownDesign(t *testing.T) {
	a := newApp(t)

	cmd := editOf(app.DesignView{
		ID:                  404,
		Name:                "Ghost",
		DefaultFilamentType: domain.PLA,
		EstimatedGrams:      grams(48),
		EstimatedMinutes:    331,
		MarginPct:           5000,
	})

	if _, err := a.EditDesign(ctx(), cmd); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("EditDesign error = %v, want ErrNotFound", err)
	}
}

func TestAddDesignRejectsMalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*app.AddDesignCmd)
		field string
	}{
		{"missing name", func(c *app.AddDesignCmd) { c.Name = "  " }, domain.FieldName},
		{"unknown filament type", func(c *app.AddDesignCmd) { c.DefaultFilamentType = "ABS" }, domain.FieldDefaultFilamentType},
		{"empty filament type", func(c *app.AddDesignCmd) { c.DefaultFilamentType = "" }, domain.FieldDefaultFilamentType},
		{"zero grams", func(c *app.AddDesignCmd) { c.EstimatedGrams = "0" }, domain.FieldEstimatedGrams},
		{"negative grams", func(c *app.AddDesignCmd) { c.EstimatedGrams = "-1" }, domain.FieldEstimatedGrams},
		{"over-precise grams", func(c *app.AddDesignCmd) { c.EstimatedGrams = "48.594" }, domain.FieldEstimatedGrams},
		{"empty grams", func(c *app.AddDesignCmd) { c.EstimatedGrams = "" }, domain.FieldEstimatedGrams},
		{"print time without a colon", func(c *app.AddDesignCmd) { c.EstimatedMinutes = "331" }, domain.FieldEstimatedMinutes},
		{"print time in words", func(c *app.AddDesignCmd) { c.EstimatedMinutes = "5h 31m" }, domain.FieldEstimatedMinutes},
		{"print time over sixty minutes", func(c *app.AddDesignCmd) { c.EstimatedMinutes = "5:61" }, domain.FieldEstimatedMinutes},
		{"zero print time", func(c *app.AddDesignCmd) { c.EstimatedMinutes = "0:00" }, domain.FieldEstimatedMinutes},
		{"empty print time", func(c *app.AddDesignCmd) { c.EstimatedMinutes = "" }, domain.FieldEstimatedMinutes},
		{"negative margin", func(c *app.AddDesignCmd) { c.MarginPct = "-1" }, domain.FieldMarginPct},
		{"malformed margin", func(c *app.AddDesignCmd) { c.MarginPct = "half" }, domain.FieldMarginPct},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			cmd := plaDesign()
			tc.mut(&cmd)

			if _, err := a.AddDesign(ctx(), cmd); err == nil {
				t.Fatal("expected AddDesign to be rejected")
			} else if msg := fieldError(t, err, tc.field); msg == "" {
				t.Errorf("no error reported against field %q: %v", tc.field, err)
			}

			designs, err := a.ListDesigns(ctx())
			if err != nil {
				t.Fatalf("ListDesigns: %v", err)
			}
			if len(designs) != 0 {
				t.Errorf("got %d designs, want none persisted", len(designs))
			}
		})
	}
}

func TestAddDesignReportsEveryFailureAtOnce(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.Name = "  "
	cmd.EstimatedMinutes = "331"

	_, err := a.AddDesign(ctx(), cmd)
	if err == nil {
		t.Fatal("expected AddDesign to be rejected")
	}
	if msg := fieldError(t, err, domain.FieldName); msg == "" {
		t.Error("no error reported against the name")
	}
	if msg := fieldError(t, err, domain.FieldEstimatedMinutes); msg == "" {
		t.Error("no error reported against the print time")
	}
}

func TestAddDesignReportsTheParseFailureNotTheInvariantItTrips(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.EstimatedGrams = "0.005"

	_, err := a.AddDesign(ctx(), cmd)
	if err == nil {
		t.Fatal("expected AddDesign to be rejected")
	}
	if msg := fieldError(t, err, domain.FieldEstimatedGrams); msg != unit.ErrMalformedGrams.Error() {
		t.Errorf("estimatedGrams error = %q, want %q", msg, unit.ErrMalformedGrams)
	}
}

func TestEditDesignRejectsMalformedInputAndKeepsTheStoredDesign(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*app.EditDesignCmd)
		field string
	}{
		{"missing name", func(c *app.EditDesignCmd) { c.Name = "" }, domain.FieldName},
		{"unknown filament type", func(c *app.EditDesignCmd) { c.DefaultFilamentType = "ABS" }, domain.FieldDefaultFilamentType},
		{"zero grams", func(c *app.EditDesignCmd) { c.EstimatedGrams = "0" }, domain.FieldEstimatedGrams},
		{"print time without a colon", func(c *app.EditDesignCmd) { c.EstimatedMinutes = "600" }, domain.FieldEstimatedMinutes},
		{"negative margin", func(c *app.EditDesignCmd) { c.MarginPct = "-5" }, domain.FieldMarginPct},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			added, err := a.AddDesign(ctx(), plaDesign())
			if err != nil {
				t.Fatalf("AddDesign: %v", err)
			}

			cmd := editOf(added)
			tc.mut(&cmd)

			if _, err := a.EditDesign(ctx(), cmd); err == nil {
				t.Fatal("expected EditDesign to be rejected")
			} else if msg := fieldError(t, err, tc.field); msg == "" {
				t.Errorf("no error reported against field %q: %v", tc.field, err)
			}

			designs, err := a.ListDesigns(ctx())
			if err != nil {
				t.Fatalf("ListDesigns: %v", err)
			}
			got := designNamed(t, designs, added.Name)
			if got != added {
				t.Errorf("design = %+v, want the rejected edit to change nothing: %+v", got, added)
			}
		})
	}
}

func TestListDesignsSurvivesAReopen(t *testing.T) {
	path := t.TempDir() + "/reopen.db"

	first := newAppAt(t, path)
	added, err := first.AddDesign(ctx(), plaDesign())
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}

	second := newAppAt(t, path)
	designs, err := second.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 1 {
		t.Fatalf("got %d designs, want 1", len(designs))
	}
	if designs[0] != added {
		t.Errorf("design = %+v, want %+v", designs[0], added)
	}
}

func TestAddDesignKeepsTheTwoDecimalsTheSlicerReports(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.EstimatedGrams = "85.59"

	design := addedDesign(t, a, cmd)
	if design.EstimatedGrams != 8559 {
		t.Errorf("EstimatedGrams = %d, want 8559", design.EstimatedGrams)
	}
}

func TestEditDesignRoundTripsAFractionalEstimate(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.EstimatedGrams = "85.59"
	design := addedDesign(t, a, cmd)

	edited, err := a.EditDesign(ctx(), editOf(design))
	if err != nil {
		t.Fatalf("EditDesign: %v", err)
	}
	if edited.EstimatedGrams != design.EstimatedGrams {
		t.Errorf("EstimatedGrams = %d, want %d", edited.EstimatedGrams, design.EstimatedGrams)
	}
}

func TestAddDesignAcceptsAHundredthOfAGram(t *testing.T) {
	a := newApp(t)

	cmd := plaDesign()
	cmd.EstimatedGrams = "0.01"

	design := addedDesign(t, a, cmd)
	if design.EstimatedGrams != 1 {
		t.Errorf("EstimatedGrams = %d, want 1", design.EstimatedGrams)
	}
}

func TestDeleteDesignRemovesItFromTheList(t *testing.T) {
	a := newApp(t)
	design := addedDesign(t, a, plaDesign())

	if err := a.DeleteDesign(ctx(), design.ID); err != nil {
		t.Fatalf("DeleteDesign: %v", err)
	}

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 0 {
		t.Errorf("ListDesigns returned %d designs, want 0", len(designs))
	}
}

func TestDeleteDesignBlockedWhenAPrintReferencesIt(t *testing.T) {
	tests := []struct {
		name   string
		prints int
		want   string
	}{
		{"one print", 1, "1 print"},
		{"three prints", 3, "3 prints"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newApp(t)
			spool := addedSpool(t, a, plaSpool())
			design := addedDesign(t, a, quotedDesign())
			for range tt.prints {
				recordedPrint(t, a, printOf(design.ID, spool.ID))
			}

			err := a.DeleteDesign(ctx(), design.ID)
			if !errors.Is(err, domain.ErrDesignPrinted) {
				t.Fatalf("DeleteDesign error = %v, want ErrDesignPrinted", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("DeleteDesign error = %q, want it to name %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "cop") {
				t.Errorf("DeleteDesign error = %q, want it to name no copy count", err)
			}

			designs, err := a.ListDesigns(ctx())
			if err != nil {
				t.Fatalf("ListDesigns: %v", err)
			}
			if len(designs) != 1 {
				t.Errorf("ListDesigns returned %d designs, want 1", len(designs))
			}
		})
	}
}

func TestDeleteDesignSucceedsOnceItsPrintIsDeleted(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	design := addedDesign(t, a, quotedDesign())
	printed := recordedPrint(t, a, printOf(design.ID, spool.ID))

	if err := a.DeletePrint(ctx(), printed.ID); err != nil {
		t.Fatalf("DeletePrint: %v", err)
	}
	if err := a.DeleteDesign(ctx(), design.ID); err != nil {
		t.Fatalf("DeleteDesign: %v", err)
	}

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 0 {
		t.Errorf("ListDesigns returned %d designs, want 0", len(designs))
	}
}

func TestDeleteDesignLeavesOtherDesignsPrintsAlone(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	printed := addedDesign(t, a, quotedDesign())
	unprinted := addedDesign(t, a, plaDesign())
	kept := recordedPrint(t, a, printOf(printed.ID, spool.ID))

	if err := a.DeleteDesign(ctx(), printed.ID); !errors.Is(err, domain.ErrDesignPrinted) {
		t.Fatalf("DeleteDesign error = %v, want ErrDesignPrinted", err)
	}
	if err := a.DeleteDesign(ctx(), unprinted.ID); err != nil {
		t.Fatalf("DeleteDesign: %v", err)
	}

	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	if len(prints) != 1 {
		t.Fatalf("ListPrints returned %d prints, want 1", len(prints))
	}
	if prints[0].ID != kept.ID {
		t.Errorf("surviving print = %d, want %d", prints[0].ID, kept.ID)
	}
}

func TestListDesignsReportsPrintCount(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	printed := addedDesign(t, a, quotedDesign())
	addedDesign(t, a, plaDesign())
	recordedPrint(t, a, printOf(printed.ID, spool.ID))
	recordedPrint(t, a, printOf(printed.ID, spool.ID))

	designs, err := a.ListDesigns(ctx())
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if got := designNamed(t, designs, "Planter").PrintCount; got != 2 {
		t.Errorf("Planter PrintCount = %d, want 2", got)
	}
	if got := designNamed(t, designs, "Cable clip").PrintCount; got != 0 {
		t.Errorf("Cable clip PrintCount = %d, want 0", got)
	}
}
