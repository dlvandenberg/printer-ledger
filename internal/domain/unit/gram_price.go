package unit

import "fmt"

// CentsPerGram is hundredths of a cent per gram: a 1000g spool bought for
// €22.00 is 220. A whole cent is too coarse for a gram price, and a Print
// freezes the price its filament cost rather than the spool it came from
// (ADR-0004). There is no Parse: the operator never types a gram price.
type CentsPerGram int64

// GramPriceScale is one cent in hundredths, the factor a CentsPerGram has to be
// divided by after it multiplies grams.
const GramPriceScale = 100

func FormatCentsPerGram(p CentsPerGram) string {
	sign := ""
	if p < 0 {
		sign, p = "-", -p
	}
	whole, frac := int64(p)/GramPriceScale, int64(p)%GramPriceScale
	switch {
	case frac == 0:
		return fmt.Sprintf("%s%d", sign, whole)
	case frac%10 == 0:
		return fmt.Sprintf("%s%d.%d", sign, whole, frac/10)
	default:
		return fmt.Sprintf("%s%d.%02d", sign, whole, frac)
	}
}
