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

type SpoolLedgers []SpoolLedger

func (l SpoolLedgers) Spools() Spools {
	spools := make(Spools, 0, len(l))
	for _, ledger := range l {
		spools = append(spools, ledger.Spool)
	}
	return spools
}

func SpoolsOf(ledgers SpoolLedgers) Spools { return ledgers.Spools() }

func (l SpoolLedgers) find(spoolID int64) (SpoolLedger, bool) {
	for _, ledger := range l {
		if ledger.Spool.ID == spoolID {
			return ledger, true
		}
	}
	return SpoolLedger{}, false
}

func (l SpoolLedgers) released(p Print) SpoolLedgers {
	released := make(SpoolLedgers, 0, len(l))
	for _, ledger := range l {
		for _, usage := range p.Usages {
			if usage.SpoolID == ledger.Spool.ID {
				ledger.UsedGrams -= usage.Grams
			}
		}
		released = append(released, ledger)
	}
	return released
}
