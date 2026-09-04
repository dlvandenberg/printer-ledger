package domain

import (
	"fmt"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	FieldDesignID  = "designID"
	FieldQuantity  = "quantity"
	FieldMinutes   = "minutes"
	FieldPrintDate = "date"
	FieldGifted    = "gifted"
	FieldKept      = "kept"
	FieldScrapped  = "scrapped"
)

// FieldUsageSpool and FieldUsageGrams key one Filament Usage row, so an error
// lands on the row that caused it rather than on the Print (ADR-0012).
func FieldUsageSpool(row int) string { return fmt.Sprintf("usage.%d.spool", row) }

func FieldUsageGrams(row int) string { return fmt.Sprintf("usage.%d.grams", row) }

type FilamentUsage struct {
	ID          int64
	SpoolID     int64
	Grams       unit.Grams
	CostPerGram unit.CentsPerGram
}

type Print struct {
	ID            int64
	DesignID      int64
	Date          time.Time
	Quantity      unit.Copies
	Minutes       unit.Minutes
	GiftedCount   unit.Copies
	KeptCount     unit.Copies
	ScrappedCount unit.Copies
	KwhPrice      unit.Cents
	KwhPerHour    unit.KwhPerHour
	MachineRate   unit.Cents
	Usages        []FilamentUsage
}

type PrintCost struct {
	Filament unit.Cents
	Energy   unit.Cents
	Overhead unit.Cents
}

// NewPrint freezes the rates a Print is costed at and then rejects what the
// operator cannot have printed. Quantity is at least 1: a run that yielded
// nothing is a Spool Adjustment (ADR-0017).
func NewPrint(p Print, s Settings, ledgers []SpoolLedger) (Print, error) {
	p = p.WithRates(s, ledgers)

	v := &ValidationError{}
	validatePrint(p, ledgers, v)
	PrintLedger{Print: p}.validateAvailable(v)

	if err := v.OrNil(); err != nil {
		return Print{}, err
	}
	return p, nil
}

// EditedPrint corrects a Print without re-costing it: the rate snapshot and the
// gram price of every Spool it already drew from are kept (ADR-0004), and only
// a row naming a Spool this Print did not use takes that Spool's price today.
// The Print's own grams are released before the overdraw check, since the
// Spool's remaining is derived from them (ADR-0003).
func EditedPrint(stored PrintLedger, p Print, ledgers []SpoolLedger) (Print, error) {
	released := releaseUsage(ledgers, stored.Print)

	p = p.WithFrozenRatesOf(stored.Print, released)
	p.ID = stored.Print.ID

	v := &ValidationError{}
	validatePrint(p, released, v)
	PrintLedger{Print: p, SoldCount: stored.SoldCount}.validateAvailable(v)

	if err := v.OrNil(); err != nil {
		return Print{}, err
	}
	return p, nil
}

func validatePrint(p Print, ledgers []SpoolLedger, v *ValidationError) {
	if p.Quantity < 1 {
		v.Add(FieldQuantity, "must be at least 1")
	}
	if p.Minutes <= 0 {
		v.Add(FieldMinutes, "must be more than 0min")
	}
	if p.Date.IsZero() {
		v.Add(FieldPrintDate, "is required")
	}
	if p.GiftedCount < 0 {
		v.Add(FieldGifted, "cannot be negative")
	}
	if p.KeptCount < 0 {
		v.Add(FieldKept, "cannot be negative")
	}
	if p.ScrappedCount < 0 {
		v.Add(FieldScrapped, "cannot be negative")
	}
	if len(p.Usages) == 0 {
		v.Add(FieldUsageSpool(0), "is required")
	}
	validateUsages(p, ledgers, v)
}

// validateUsages sums grams per Spool rather than checking them per row
// (ADR-0021), and reports the overdraw once, on the row that crosses what is
// left.
func validateUsages(p Print, ledgers []SpoolLedger, v *ValidationError) {
	printType := p.FilamentType(ledgers)
	requested := map[int64]unit.Grams{}
	overdrawn := map[int64]bool{}
	for row, usage := range p.Usages {
		ledger, ok := findLedger(ledgers, usage.SpoolID)
		switch {
		case !ok:
			v.Add(FieldUsageSpool(row), "is not a spool on the shelf")
		case ledger.Spool.FilamentType != printType:
			v.Add(FieldUsageSpool(row), fmt.Sprintf("cannot mix %s with %s on one print", ledger.Spool.FilamentType, printType))
		case usage.Grams <= 0:
			v.Add(FieldUsageGrams(row), "must be more than 0g")
		default:
			requested[usage.SpoolID] += usage.Grams
			if requested[usage.SpoolID] > ledger.Remaining() && !overdrawn[usage.SpoolID] {
				overdrawn[usage.SpoolID] = true
				v.Add(FieldUsageGrams(row), fmt.Sprintf("only %s left", unit.FormatGrams(ledger.Remaining())))
			}
		}
	}
}

// WithRates copies the rates in force and each Spool's gram price onto the
// Print, so no later edit to Settings or to a Spool rewrites what it cost
// (ADR-0004). It validates nothing, so a half-typed form can still be costed.
func (p Print) WithRates(s Settings, ledgers []SpoolLedger) Print {
	usages := make([]FilamentUsage, 0, len(p.Usages))
	for _, usage := range p.Usages {
		if ledger, ok := findLedger(ledgers, usage.SpoolID); ok {
			usage.CostPerGram = ledger.Spool.GramPrice()
		}
		usages = append(usages, usage)
	}
	p.Usages = usages
	p.KwhPrice = s.KwhPrice
	p.KwhPerHour = s.PowerRate(p.FilamentType(ledgers)).KwhPerHour
	p.MachineRate = s.MachineHourlyRate
	return p
}

// WithFrozenRatesOf costs an edited Print the way the stored one was costed:
// the rate snapshot stands, and so does the gram price of every Spool it
// already drew from. A row naming a Spool it did not use takes that Spool's
// price today (ADR-0004).
func (p Print) WithFrozenRatesOf(stored Print, ledgers []SpoolLedger) Print {
	p.KwhPrice = stored.KwhPrice
	p.KwhPerHour = stored.KwhPerHour
	p.MachineRate = stored.MachineRate
	p.Usages = repriced(p.Usages, stored, ledgers)
	return p
}

func (p Print) Cost() PrintCost {
	return PrintCost{
		Filament: p.filamentCost(),
		Energy:   energyCost(p.Minutes, p.KwhPerHour, p.KwhPrice),
		Overhead: overheadCost(p.Minutes, p.MachineRate),
	}
}

func (p Print) UsedGrams() unit.Grams {
	var grams unit.Grams
	for _, usage := range p.Usages {
		grams += usage.Grams
	}
	return grams
}

func (c PrintCost) JobCost() unit.Cents { return c.Filament + c.Energy + c.Overhead }

// CostPerCopy truncates: the sub-cent per copy is accepted, and Suggested Price
// rounds up to €0.50 anyway (ADR-0002, ADR-0007).
func (c PrintCost) CostPerCopy(quantity unit.Copies) unit.Cents {
	if quantity < 1 {
		return 0
	}
	return c.JobCost() / unit.Cents(quantity)
}

// filamentCost sums in hundredths of a cent and rounds once, so a gram price
// that is not a whole cent does not round on every row (ADR-0002).
func (p Print) filamentCost() unit.Cents {
	var hundredths int64
	for _, usage := range p.Usages {
		hundredths += int64(usage.Grams) * int64(usage.CostPerGram)
	}
	return unit.Cents(roundDiv(hundredths, unit.GramPriceScale))
}

// FilamentType is the type the energy rate is looked up by. All usage rows are
// Spools of one Filament Type (ADR-0012), so the first row that names a spool
// on the shelf decides, and every later row is validated against it.
func (p Print) FilamentType(ledgers []SpoolLedger) FilamentType {
	for _, usage := range p.Usages {
		if ledger, ok := findLedger(ledgers, usage.SpoolID); ok {
			return ledger.Spool.FilamentType
		}
	}
	return ""
}

func repriced(usages []FilamentUsage, stored Print, ledgers []SpoolLedger) []FilamentUsage {
	frozen := map[int64]unit.CentsPerGram{}
	for _, usage := range stored.Usages {
		frozen[usage.SpoolID] = usage.CostPerGram
	}

	priced := make([]FilamentUsage, 0, len(usages))
	for _, usage := range usages {
		if price, ok := frozen[usage.SpoolID]; ok {
			usage.CostPerGram = price
		} else if ledger, found := findLedger(ledgers, usage.SpoolID); found {
			usage.CostPerGram = ledger.Spool.GramPrice()
		}
		priced = append(priced, usage)
	}
	return priced
}

func releaseUsage(ledgers []SpoolLedger, p Print) []SpoolLedger {
	released := make([]SpoolLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		for _, usage := range p.Usages {
			if usage.SpoolID == ledger.Spool.ID {
				ledger.UsedGrams -= usage.Grams
			}
		}
		released = append(released, ledger)
	}
	return released
}

func findLedger(ledgers []SpoolLedger, spoolID int64) (SpoolLedger, bool) {
	for _, ledger := range ledgers {
		if ledger.Spool.ID == spoolID {
			return ledger, true
		}
	}
	return SpoolLedger{}, false
}
