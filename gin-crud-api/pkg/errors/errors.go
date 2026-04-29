package errors

import (
	"errors"
	"net/http"
)

// AppError represents a structured error for HTTP APIs.
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ErrInternalServer is a default error for unexpected failures.
var ErrInternalServer = &AppError{
	Code:    "INTERNAL_SERVER_ERROR",
	Message: "internal server error",
	Status:  http.StatusInternalServerError,
}

// ErrDuplicateEmail is returned when a user tries to create an account with an email that already exists.
var ErrDuplicateEmail = &AppError{
	Code:    "DUPLICATE_EMAIL",
	Message: "email already exists",
	Status:  http.StatusBadRequest,
}

// ErrDuplicateUsername is returned when a user tries to create an account with a username that already exists.
var ErrDuplicateUsername = &AppError{
	Code:    "DUPLICATE_USERNAME",
	Message: "username already exists",
	Status:  http.StatusBadRequest,
}

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = &AppError{
	Code:    "USER_NOT_FOUND",
	Message: "user not found",
	Status:  http.StatusNotFound,
}

// ErrInvalidRequest is returned when request body validation fails.
var ErrInvalidRequest = &AppError{
	Code:    "INVALID_REQUEST",
	Message: "invalid request",
	Status:  http.StatusBadRequest,
}

// ErrrInvalidUserId is returned when the provided user ID is not a valid UUID.
var ErrInvalidUserId = &AppError{
	Code:    "INVALID_USER_ID",
	Message: "invalid user ID",
	Status:  http.StatusBadRequest,
}

// GetError normalizes any error into an AppError.
func GetError(err error) *AppError {
	if err == nil {
		return ErrInternalServer
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return NewInternalServer(err)
}

// NewInternalServer wraps the original error for logging or tracing.
func NewInternalServer(err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}
