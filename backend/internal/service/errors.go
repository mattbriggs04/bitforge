package service

import "errors"

type ErrorKind string

const (
	ErrorKindInvalid     ErrorKind = "invalid"
	ErrorKindNotFound    ErrorKind = "not_found"
	ErrorKindUnsupported ErrorKind = "unsupported"
	ErrorKindForbidden   ErrorKind = "forbidden"
	ErrorKindConflict    ErrorKind = "conflict"
)

type AppError struct {
	Kind    ErrorKind
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func newInvalidError(message string) error {
	return &AppError{Kind: ErrorKindInvalid, Message: message}
}

func newNotFoundError(message string) error {
	return &AppError{Kind: ErrorKindNotFound, Message: message}
}

func newUnsupportedError(message string) error {
	return &AppError{Kind: ErrorKindUnsupported, Message: message}
}

func newForbiddenError(message string) error {
	return &AppError{Kind: ErrorKindForbidden, Message: message}
}

func newConflictError(message string) error {
	return &AppError{Kind: ErrorKindConflict, Message: message}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
