package httperrors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestIsHTTPCodeError(t *testing.T) {
	assert.True(t, httperrors.IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusBadRequest))
	assert.False(t, httperrors.IsHTTPCodeError(nil, http.StatusBadRequest))
	assert.False(t, httperrors.IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusNotFound))
	assert.False(t, httperrors.IsHTTPCodeError(errors.New("not an http error"), http.StatusNotFound))
}

func TestIs404Error(t *testing.T) {
	assert.True(t, httperrors.Is404(&scw.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, httperrors.Is404(nil))
	assert.False(t, httperrors.Is404(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIs403Error(t *testing.T) {
	assert.True(t, httperrors.Is403(&scw.ResponseError{StatusCode: http.StatusForbidden}))
	assert.False(t, httperrors.Is403(nil))
	assert.False(t, httperrors.Is403(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIsStatusCodeError(t *testing.T) {
	// ResponseError matches on its status code.
	assert.True(t, httperrors.IsStatusCodeError(&scw.ResponseError{StatusCode: http.StatusNotFound}, http.StatusNotFound))
	assert.True(t, httperrors.IsStatusCodeError(&scw.ResponseError{StatusCode: http.StatusNotFound}, http.StatusForbidden, http.StatusNotFound))
	assert.False(t, httperrors.IsStatusCodeError(&scw.ResponseError{StatusCode: http.StatusNotFound}, http.StatusForbidden))
	assert.False(t, httperrors.IsStatusCodeError(nil, http.StatusNotFound))
	assert.False(t, httperrors.IsStatusCodeError(errors.New("not an http error"), http.StatusNotFound))
	assert.False(t, httperrors.IsStatusCodeError(&scw.ResponseError{StatusCode: http.StatusNotFound}))

	// SDK typed errors are matched against the status code they correspond to.
	assert.True(t, httperrors.IsStatusCodeError(&scw.PermissionsDeniedError{}, http.StatusForbidden))
	assert.True(t, httperrors.IsStatusCodeError(&scw.ResourceNotFoundError{}, http.StatusNotFound))
	assert.True(t, httperrors.IsStatusCodeError(&scw.TransientStateError{}, http.StatusConflict))
	assert.True(t, httperrors.IsStatusCodeError(&scw.PreconditionFailedError{}, http.StatusPreconditionFailed))
	assert.False(t, httperrors.IsStatusCodeError(&scw.PermissionsDeniedError{}, http.StatusNotFound))
}
