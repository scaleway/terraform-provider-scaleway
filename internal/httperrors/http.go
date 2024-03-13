package httperrors

import (
	"errors"
	"net/http"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// IsHTTPCodeError returns true if err is an http error with code statusCode
func IsHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if errors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}
	return false
}

// Is404 returns true if err is an HTTP 404 error
func Is404(err error) bool {
	notFoundError := &scw.ResourceNotFoundError{}
	return IsHTTPCodeError(err, http.StatusNotFound) || errors.As(err, &notFoundError)
}

func Is412(err error) bool {
	preConditionFailedError := &scw.PreconditionFailedError{}
	return IsHTTPCodeError(err, http.StatusPreconditionFailed) || errors.As(err, &preConditionFailedError)
}

// Is403 returns true if err is an HTTP 403 error
func Is403(err error) bool {
	permissionsDeniedError := &scw.PermissionsDeniedError{}
	return IsHTTPCodeError(err, http.StatusForbidden) || errors.As(err, &permissionsDeniedError)
}

// Is409 return true is err is an HTTP 409 error
func Is409(err error) bool {
	// check transient error
	transientStateError := &scw.TransientStateError{}
	return IsHTTPCodeError(err, http.StatusConflict) || errors.As(err, &transientStateError)
}

// Is410 returns true if err is an HTTP 410 error
func Is410(err error) bool {
	return IsHTTPCodeError(err, http.StatusGone)
}
