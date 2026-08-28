package domain

const (
	FieldKwhPrice            = "kwhPrice"
	FieldMachineHourlyRate   = "machineHourlyRate"
	FieldPrinterPurchaseCost = "printerPurchaseCost"
	FieldDefaultMargin       = "defaultMargin"
	FieldMinMargin           = "minMargin"
)

// FieldPowerRate keys one rate row per FilamentType, so adding a material adds
// a row rather than a constant (ADR-0010).
func FieldPowerRate(t FilamentType) string { return "powerRate." + t.String() }

type RateSource string

const (
	RateDefault  RateSource = "default"
	RateMeasured RateSource = "measured"
)

type PowerRate struct {
	FilamentType FilamentType
	KwhPerHour   KwhPerHour
	Measured     bool
}

func (r PowerRate) Source() RateSource {
	if r.Measured {
		return RateMeasured
	}
	return RateDefault
}

type Settings struct {
	KwhPrice            Cents
	MachineHourlyRate   Cents
	PrinterPurchaseCost Cents
	DefaultMargin       Percent
	MinMargin           Percent
	PowerRates          []PowerRate
}

// DefaultSettings seeds a fresh ledger, so energy is never silently costed at
// zero before anything has been measured (ADR-0008).
func DefaultSettings() Settings {
	return Settings{
		KwhPrice:            28,
		MachineHourlyRate:   35,
		PrinterPurchaseCost: 0,
		DefaultMargin:       5000,
		MinMargin:           1500,
		PowerRates: []PowerRate{
			{FilamentType: PLA, KwhPerHour: 0.09},
			{FilamentType: PLAPlus, KwhPerHour: 0.10},
			{FilamentType: PETG, KwhPerHour: 0.12},
		},
	}
}

func NewSettings(s Settings) (Settings, error) {
	v := &ValidationError{}
	if s.KwhPrice < 0 {
		v.Add(FieldKwhPrice, "cannot be negative")
	}
	if s.MachineHourlyRate < 0 {
		v.Add(FieldMachineHourlyRate, "cannot be negative")
	}
	if s.PrinterPurchaseCost < 0 {
		v.Add(FieldPrinterPurchaseCost, "cannot be negative")
	}
	if s.DefaultMargin < 0 {
		v.Add(FieldDefaultMargin, "cannot be negative")
	}
	if s.MinMargin < 0 {
		v.Add(FieldMinMargin, "cannot be negative")
	}

	rates := make([]PowerRate, 0, len(FilamentTypes()))
	for _, t := range FilamentTypes() {
		rate, ok := findPowerRate(s.PowerRates, t)
		switch {
		case !ok:
			v.Add(FieldPowerRate(t), "is required")
		case rate.KwhPerHour <= 0:
			v.Add(FieldPowerRate(t), "must be more than 0")
		}
		rate.FilamentType = t
		rates = append(rates, rate)
	}
	s.PowerRates = rates

	if err := v.OrNil(); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func (s Settings) PowerRate(t FilamentType) PowerRate {
	rate, _ := findPowerRate(s.PowerRates, t)
	rate.FilamentType = t
	return rate
}

// WithPowerRate marks a rate measured as soon as its value moves off what the
// ledger held, which is the only signal that the operator has put a smart plug
// on the printer.
func (s Settings) WithPowerRate(t FilamentType, k KwhPerHour) Settings {
	current := s.PowerRate(t)
	updated := PowerRate{FilamentType: t, KwhPerHour: k, Measured: current.Measured || k != current.KwhPerHour}

	rates := make([]PowerRate, 0, len(s.PowerRates))
	replaced := false
	for _, rate := range s.PowerRates {
		if rate.FilamentType == t {
			rate, replaced = updated, true
		}
		rates = append(rates, rate)
	}
	if !replaced {
		rates = append(rates, updated)
	}
	s.PowerRates = rates
	return s
}

func findPowerRate(rates []PowerRate, t FilamentType) (PowerRate, bool) {
	for _, rate := range rates {
		if rate.FilamentType == t {
			return rate, true
		}
	}
	return PowerRate{}, false
}
