package domain

type SpoolLedger struct {
	Spool         Spool
	UsedGrams     Grams
	AdjustedGrams Grams
}

func (l SpoolLedger) Remaining() Grams {
	return l.Spool.InitialGrams - l.UsedGrams + l.AdjustedGrams
}

func (l SpoolLedger) RemainingValue() Cents {
	return l.Spool.ValueOf(l.Remaining())
}

func (l SpoolLedger) State() SpoolState {
	if l.Remaining() > 0 {
		return SpoolActive
	}
	return SpoolEmpty
}
