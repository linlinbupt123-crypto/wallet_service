package errors

import "fmt"

type AppError struct {
	Code Code
	Op   string
	Err  error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Op, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func WrapWithCode(code Code, op string, err error) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Code: code,
		Op:   op,
		Err:  err,
	}
}
