package app

import (
	"errors"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func parseField[T any](errs *domain.ValidationError, name, raw string, parse func(string) (T, error)) T {
	v, err := parse(raw)
	if err != nil {
		errs.Add(name, err.Error())
		var zero T
		return zero
	}
	return v
}

func fallback[T any](def T, parse func(string) (T, error)) func(string) (T, error) {
	return func(raw string) (T, error) {
		if strings.TrimSpace(raw) == "" {
			return def, nil
		}
		return parse(raw)
	}
}

func validate[T any](value T, errs *domain.ValidationError, invariants func(T) (T, error)) (T, error) {
	valid, err := invariants(value)
	if err != nil {
		var v *domain.ValidationError
		if !errors.As(err, &v) {
			var zero T
			return zero, err
		}
		errs.MergeMissing(v)
	}
	if err := errs.OrNil(); err != nil {
		var zero T
		return zero, err
	}
	return valid, nil
}
