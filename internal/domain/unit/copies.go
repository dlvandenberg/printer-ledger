package unit

import (
	"errors"
	"strconv"
	"strings"
)

// Copies counts physical printed objects. Copies are counted, never
// individually recorded (ADR-0013).
type Copies int64

var ErrMalformedCopies = errors.New("not a whole number of copies")

// ParseCopies rejects a negative count as malformed: no field counts copies
// backwards. Zero parses — the "at least 1" rule is Quantity's alone.
func ParseCopies(s string) (Copies, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "x")
	s = strings.TrimSpace(s)

	copies, err := strconv.ParseInt(s, 10, 64)
	if err != nil || copies < 0 {
		return 0, ErrMalformedCopies
	}
	return Copies(copies), nil
}

func FormatCopies(c Copies) string { return strconv.FormatInt(int64(c), 10) }
