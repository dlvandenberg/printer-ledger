package domain

import "github.com/dlvandenberg/printer-ledger/internal/domain/unit"

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
