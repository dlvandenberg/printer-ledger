package domain

import (
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

var ErrSpoolDrawnFrom = errors.New("cannot delete a spool a print has drawn from")

type SpoolLedger struct {
	Spool         Spool
	UsedGrams     unit.Grams
	AdjustedGrams unit.Grams
}

func SpoolsOf(ledgers []SpoolLedger) []Spool {
	spools := make([]Spool, 0, len(ledgers))
	for _, ledger := range ledgers {
		spools = append(spools, ledger.Spool)
	}
	return spools
}

func (l SpoolLedger) Remaining() unit.Grams {
	return l.Spool.InitialGrams - l.UsedGrams + l.AdjustedGrams
}

func (l SpoolLedger) RemainingValue() unit.Cents {
	return l.Spool.ValueOf(l.Remaining())
}

func (l SpoolLedger) State() SpoolState {
	if l.Remaining() > 0 {
		return SpoolActive
	}
	return SpoolEmpty
}

func (l SpoolLedger) DeleteBlocked() error {
	if l.UsedGrams == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s used", ErrSpoolDrawnFrom, unit.FormatGrams(l.UsedGrams))
}
