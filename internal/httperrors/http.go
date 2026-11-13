package httperrors

import (
	"errors"
	"net/http"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// IsHTTPCodeError returns true if err is an http error with code statusCode
func IsHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	if responseError, ok := errors.AsType[*scw.ResponseError](err); ok && responseError.StatusCode == statusCode {
		return true
	}

	return false
}

// Map SDK typed errors to http errors.
func StatusCodeOf(err error) int {
	var (
		responseError           *scw.ResponseError
		permissionsDeniedError  *scw.PermissionsDeniedError
		resourceNotFoundError   *scw.ResourceNotFoundError
		transientStateError     *scw.TransientStateError
		preconditionFailedError *scw.PreconditionFailedError
	)

	switch {
	case errors.As(err, &responseError):
		return responseError.StatusCode
	case errors.As(err, &permissionsDeniedError):
		return http.StatusForbidden
	case errors.As(err, &resourceNotFoundError):
		return http.StatusNotFound
	case errors.As(err, &transientStateError):
		return http.StatusConflict
	case errors.As(err, &preconditionFailedError):
		return http.StatusPreconditionFailed
	default:
		return 0
	}
}

// IsStatusCodeError returns true if err is an HTTP error with one of the given status codes.
func IsStatusCodeError(err error, statusCodes ...int) bool {
	if err == nil {
		return false
	}

	code := StatusCodeOf(err)

	return slices.Contains(statusCodes, code)
}

// Is404 returns true if err is an HTTP 404 error
func Is404(err error) bool {
	return IsStatusCodeError(err, http.StatusNotFound)
}

func Is412(err error) bool {
	return IsStatusCodeError(err, http.StatusPreconditionFailed)
}

// Is403 returns true if err is an HTTP 403 error
func Is403(err error) bool {
	return IsStatusCodeError(err, http.StatusForbidden)
}

// Is409 return true is err is an HTTP 409 error
func Is409(err error) bool {
	return IsStatusCodeError(err, http.StatusConflict)
}

// Is410 returns true if err is an HTTP 410 error
func Is410(err error) bool {
	return IsStatusCodeError(err, http.StatusGone)
}
