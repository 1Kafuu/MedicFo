package helper

import (
	"errors"
	"net/http"
)

// AppError: error domain yang tahu kode & status HTTP-nya.
type AppError struct {
	Code   string
	Status int
	Detail string
}

func (e *AppError) Error() string { return e.Code + ": " + e.Detail }

// Sentinel error domain — daftar kode harus sama dengan API.md.
var (
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Status: http.StatusUnauthorized, Detail: "email or password is wrong"}
	ErrRateLimited        = &AppError{Code: "RATE_LIMITED", Status: http.StatusTooManyRequests, Detail: "too many attempts, try again later"}
	ErrVersionConflict    = &AppError{Code: "VERSION_CONFLICT", Status: http.StatusConflict, Detail: "data was modified, reload and retry"}
	ErrQuotaExceeded      = &AppError{Code: "QUOTA_EXCEEDED", Status: http.StatusConflict, Detail: "doctor schedule quota is full"}
	ErrForbidden          = &AppError{Code: "FORBIDDEN", Status: http.StatusForbidden, Detail: "you do not have access to this resource"}
	ErrDuplicateReview    = &AppError{Code: "DUPLICATE_REVIEW", Status: http.StatusConflict, Detail: "you already reviewed this doctor"}
	ErrNotFound           = &AppError{Code: "NOT_FOUND", Status: http.StatusNotFound, Detail: "resource not found"}
	ErrTokenExpired       = &AppError{Code: "TOKEN_EXPIRED", Status: http.StatusUnauthorized, Detail: "token has expired"}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Status: http.StatusUnauthorized, Detail: "authentication required"}
	ErrInternal           = &AppError{Code: "INTERNAL_ERROR", Status: http.StatusInternalServerError, Detail: "internal server error"}
)

// AsAppError mengubah error apa pun menjadi *AppError.
// Error yang bukan *AppError dianggap internal (fail-closed, fallback 500).
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return ErrInternal
}
