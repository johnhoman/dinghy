package codec

import "fmt"

func ErrRequiredField(name string) error {
	return errRequiredField(name)
}

type errRequiredField string

func (err errRequiredField) Error() string {
	return fmt.Sprintf("missing required field: %q", string(err))
}

func ErrUnexpectedField(name string) error {
	return errUnexpectedField(name)
}

type errUnexpectedField string

func (err errUnexpectedField) Error() string {
	return fmt.Sprintf("unexpected field: %q", string(err))
}
