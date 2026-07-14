package consumer

import "errors"

type permanentError struct {
	err error
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err: err}
}

func IsPermanent(err error) bool {
	var permanentErr permanentError

	return errors.As(err, &permanentErr)
}

func (e permanentError) Error() string {
	return e.err.Error()
}

func (e permanentError) Unwrap() error {
	return e.err
}
