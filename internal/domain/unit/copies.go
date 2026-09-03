package unit

import (
	"errors"
	"strconv"
	"strings"
)

// Copies counts physical printed objects. Copies are counted, never
// individually recorded (ADR-0013).
type Copies int

var ErrMalformedCopies = errors.New("not a whole number of copies")

func ParseCopies(s string) (Copies, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "x")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrMalformedCopies
	}
	copies, err := strconv.Atoi(s)
	if err != nil {
		return 0, ErrMalformedCopies
	}
	return Copies(copies), nil
}

func FormatCopies(c Copies) string { return strconv.Itoa(int(c)) }
