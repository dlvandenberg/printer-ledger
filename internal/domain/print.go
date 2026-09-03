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
	ID          int64
	DesignID    int64
	Date        time.Time
	Quantity    unit.Copies
	Minutes     unit.Minutes
	KwhPrice    unit.Cents
	KwhPerHour  unit.KwhPerHour
	MachineRate unit.Cents
	Usages      []FilamentUsage
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
	if p.Quantity < 1 {
		v.Add(FieldQuantity, "must be at least 1")
	}
	if p.Minutes <= 0 {
		v.Add(FieldMinutes, "must be more than 0min")
	}
	if p.Date.IsZero() {
		v.Add(FieldPrintDate, "is required")
	}
	if len(p.Usages) == 0 {
		v.Add(FieldUsageSpool(0), "is required")
	}
	for row, usage := range p.Usages {
		ledger, ok := findLedger(ledgers, usage.SpoolID)
		switch {
		case !ok:
			v.Add(FieldUsageSpool(row), "is not a spool on the shelf")
		case usage.Grams <= 0:
			v.Add(FieldUsageGrams(row), "must be more than 0g")
		case usage.Grams > ledger.Remaining():
			v.Add(FieldUsageGrams(row), fmt.Sprintf("only %s left", unit.FormatGrams(ledger.Remaining())))
		}
	}

	if err := v.OrNil(); err != nil {
		return Print{}, err
	}
	return p, nil
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
	p.KwhPerHour = s.PowerRate(p.filamentType(ledgers)).KwhPerHour
	p.MachineRate = s.MachineHourlyRate
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

// CostPerCopy rounds where the rest of the cost math truncates: half a cent per
// copy is money the job really spent, and the pinned reference case is 236c of
// 471c over two copies.
func (c PrintCost) CostPerCopy(quantity unit.Copies) unit.Cents {
	if quantity < 1 {
		return 0
	}
	return unit.Cents(roundDiv(int64(c.JobCost()), int64(quantity)))
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

// filamentType is the type the energy rate is looked up by. All usage rows are
// Spools of one Filament Type (ADR-0012), so the first row that names a spool
// on the shelf decides.
func (p Print) filamentType(ledgers []SpoolLedger) FilamentType {
	for _, usage := range p.Usages {
		if ledger, ok := findLedger(ledgers, usage.SpoolID); ok {
			return ledger.Spool.FilamentType
		}
	}
	return ""
}

func findLedger(ledgers []SpoolLedger, spoolID int64) (SpoolLedger, bool) {
	for _, ledger := range ledgers {
		if ledger.Spool.ID == spoolID {
			return ledger, true
		}
	}
	return SpoolLedger{}, false
}
