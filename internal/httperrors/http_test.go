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
