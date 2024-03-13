package errs_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/stretchr/testify/assert"
)

func TestIsHTTPCodeError(t *testing.T) {
	assert.True(t, errs.IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusBadRequest))
	assert.False(t, errs.IsHTTPCodeError(nil, http.StatusBadRequest))
	assert.False(t, errs.IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusNotFound))
	assert.False(t, errs.IsHTTPCodeError(errors.New("not an http error"), http.StatusNotFound))
}

func TestIs404Error(t *testing.T) {
	assert.True(t, errs.Is404Error(&scw.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, errs.Is404Error(nil))
	assert.False(t, errs.Is404Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIs403Error(t *testing.T) {
	assert.True(t, errs.Is403Error(&scw.ResponseError{StatusCode: http.StatusForbidden}))
	assert.False(t, errs.Is403Error(nil))
	assert.False(t, errs.Is403Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}
