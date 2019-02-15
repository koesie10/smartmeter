package smartmeter

import "fmt"

type ParseError struct {
	Value string
	Type  string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %v as %s: %v", e.Value, e.Type, e.Err)
}

func WrapError(err error, t, value string) error {
	return &ParseError{
		Value: value,
		Type:  t,
		Err:   err,
	}
}
