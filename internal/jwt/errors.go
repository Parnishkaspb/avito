package jwt

import "fmt"

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrInvalidToken         = &Error{Code: "INVALID_TOKEN", Message: "Token is invalid"}
	ErrExpiredToken         = &Error{Code: "EXPIRED_TOKEN", Message: "Token has expired"}
	ErrInvalidSigningMethod = &Error{Code: "INVALID_SIGNING_METHOD", Message: "Unexpected signing method"}
	ErrTokenGeneration      = &Error{Code: "TOKEN_GENERATION_FAILED", Message: "Failed to generate token"}
)
