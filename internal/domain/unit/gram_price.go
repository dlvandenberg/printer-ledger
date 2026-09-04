package unit

// CentsPerGram is hundredths of a cent per whole gram: a 1000g spool bought for
// €22.00 is 220. A whole cent is too coarse for a gram price, and a Print
// freezes the price its filament cost rather than the spool it came from
// (ADR-0004). There is no Parse: the operator never types a gram price.
type CentsPerGram int64

// GramPriceScale is one cent in hundredths, the factor a CentsPerGram has to be
// divided by after it multiplies a weight.
const GramPriceScale = 100

func FormatCentsPerGram(p CentsPerGram) string { return formatHundredths(int64(p), "") }
