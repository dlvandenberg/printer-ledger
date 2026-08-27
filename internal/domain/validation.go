package domain

import "strings"

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	msgs := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		msgs = append(msgs, f.Field+": "+f.Message)
	}
	return strings.Join(msgs, "; ")
}

func (e *ValidationError) For(field string) string {
	if e == nil {
		return ""
	}
	for _, f := range e.Fields {
		if f.Field == field {
			return f.Message
		}
	}
	return ""
}

func (e *ValidationError) HasError() bool { return e != nil && len(e.Fields) > 0 }

func (e *ValidationError) Add(field, message string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
}

func (e *ValidationError) OrNil() error {
	if e.HasError() {
		return e
	}
	return nil
}
