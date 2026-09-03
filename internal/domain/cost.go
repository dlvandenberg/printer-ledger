package domain

import "github.com/dlvandenberg/printer-ledger/internal/domain/unit"

// energyCost scales the rate to an integer before it multiplies money, so a
// half-cent result does not depend on how a float64 happens to hold the rate.
func energyCost(m unit.Minutes, k unit.KwhPerHour, price unit.Cents) unit.Cents {
	const rateScale = 1_000_000
	rate := int64(float64(k)*rateScale + 0.5)
	return unit.Cents(roundDiv(int64(m)*rate*int64(price), unit.MIN_IN_HOUR*rateScale))
}

func overheadCost(m unit.Minutes, rate unit.Cents) unit.Cents {
	return unit.Cents(roundDiv(int64(m)*int64(rate), unit.MIN_IN_HOUR))
}

func roundDiv(a, b int64) int64 { return (a + b/2) / b }

func ceilDiv(a, b int64) int64 { return (a + b - 1) / b }
