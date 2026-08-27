package domain

type FilamentType string

const (
	PLA     FilamentType = "PLA"
	PLAPlus FilamentType = "PLA+"
	PETG    FilamentType = "PETG"
)

func FilamentTypes() []FilamentType { return []FilamentType{PLA, PLAPlus, PETG} }

func (t FilamentType) Valid() bool {
	for _, known := range FilamentTypes() {
		if t == known {
			return true
		}
	}
	return false
}

func (t FilamentType) String() string { return string(t) }
