package models

import "fmt"

type HTTPError struct {
	Code    int    `json:"-"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e HTTPError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type FormatValidationError struct {
	ValidationError
}

type DataValidationError struct {
	ValidationError
}

type InvalidJSONError struct {
	ValidationError
}

type NotFoundError struct {
	ValidationError
}

func (e NotFoundError) Error() string {
	if e.Message == "" {
		return "resouce not found"
	}
	return e.Message
}

// WrapError wraps a plain error into a custom error
func WrapError(customErr string, originalErr error) error {
	err := fmt.Errorf("%s: %v", customErr, originalErr)
	return err
}
