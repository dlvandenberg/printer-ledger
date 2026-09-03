package domain

import "github.com/dlvandenberg/printer-ledger/internal/domain/unit"

const priceStepCents = 50

// EstimatedCost is what one copy of a Design would cost in one Filament Type,
// priced against the most expensive Capable Spool of that type (ADR-0015).
type EstimatedCost struct {
	FilamentType FilamentType
	Spool        Spool
	// Reference marks a row priced off a spool that cannot supply the job,
	// because no Capable Spool of that type is in stock.
	Reference bool
	Filament  unit.Cents
	Energy    unit.Cents
	Overhead  unit.Cents
}

func (e EstimatedCost) Total() unit.Cents { return e.Filament + e.Energy + e.Overhead }

// DesignQuote is a Design's Estimated Cost in every Filament Type stocked plus
// the one Suggested Price to say out loud. It reads current Settings, so it
// moves with rates and stock.
type DesignQuote struct {
	Design Design
	Costs  []EstimatedCost
}

func NewDesignQuote(d Design, s Settings, ledgers []SpoolLedger) DesignQuote {
	costs := make([]EstimatedCost, 0, len(FilamentTypes()))
	for _, t := range FilamentTypes() {
		spool, reference, stocked := priceSpool(t, d.EstimatedGrams, ledgers)
		if !stocked {
			continue
		}
		costs = append(costs, EstimatedCost{
			FilamentType: t,
			Spool:        spool,
			Reference:    reference,
			Filament:     spool.QuotedValueOf(d.EstimatedGrams),
			Energy:       energyCost(d.EstimatedMinutes, s.PowerRate(t).KwhPerHour, s.KwhPrice),
			Overhead:     overheadCost(d.EstimatedMinutes, s.MachineHourlyRate),
		})
	}
	return DesignQuote{Design: d, Costs: costs}
}

func (q DesignQuote) DefaultCost() (EstimatedCost, bool) {
	for _, cost := range q.Costs {
		if cost.FilamentType == q.Design.DefaultFilamentType {
			return cost, true
		}
	}
	return EstimatedCost{}, false
}

func (q DesignQuote) SuggestedPrice() (unit.Cents, bool) {
	cost, ok := q.DefaultCost()
	if !ok {
		return 0, false
	}
	return SuggestedPrice(cost.Total(), q.Design.MarginPct), true
}

// SuggestedPrice rounds up to the nearest 50 cents, never to the nearest, so
// the price never falls below the intended margin (ADR-0007).
func SuggestedPrice(cost unit.Cents, margin unit.Percent) unit.Cents {
	scaled := int64(cost) * int64(unit.PercentScale+margin)
	step := int64(priceStepCents) * int64(unit.PercentScale)
	return unit.Cents(ceilDiv(scaled, step) * priceStepCents)
}

// priceSpool picks the spool a Filament Type row is priced against: the most
// expensive Capable Spool, else the most expensive of that type ever bought as
// a reference price, else nothing and no row at all (ADR-0015).
func priceSpool(t FilamentType, grams unit.Grams, ledgers []SpoolLedger) (spool Spool, reference, stocked bool) {
	var capable, everBought Spool
	var haveCapable bool
	for _, ledger := range ledgers {
		if ledger.Spool.FilamentType != t {
			continue
		}
		if !stocked || ledger.Spool.PricierPerGramThan(everBought) {
			everBought, stocked = ledger.Spool, true
		}
		if ledger.Remaining() < grams {
			continue
		}
		if !haveCapable || ledger.Spool.PricierPerGramThan(capable) {
			capable, haveCapable = ledger.Spool, true
		}
	}
	if haveCapable {
		return capable, false, true
	}
	return everBought, stocked, stocked
}
