package verify_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
	"github.com/stretchr/testify/assert"
)

func TestValidationUUIDWithInvalidUUIDReturnError(t *testing.T) {
	assert := assert.New(t)

	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid"} {
		warnings, errors := verify.IsUUID()(uuid, "key")
		assert.Empty(warnings)
		assert.Len(errors, 1)
	}
}

func TestValidationUUIDWithValidUUIDReturnNothing(t *testing.T) {
	assert := assert.New(t)

	warnings, errors := verify.IsUUID()("6ba7b810-9dad-11d1-80b4-00c04fd430c8", "key")

	assert.Empty(warnings)
	assert.Empty(errors)
}

func TestValidationUUIDorUUIDWithLocalityWithValidUUIDReturnNothing(t *testing.T) {
	assert := assert.New(t)

	for _, uuid := range []string{"fr-par/6ba7b810-9dad-11d1-80b4-00c04fd430c8", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		warnings, errors := verify.IsUUIDorUUIDWithLocality()(uuid, "key")
		assert.Empty(warnings)
		assert.Empty(errors)
	}
}

func TestValidationUUIDorUUIDWithLocalityWithInvalidUUIDReturnError(t *testing.T) {
	assert := assert.New(t)

	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid"} {
		warnings, errors := verify.IsUUIDorUUIDWithLocality()(uuid, "key")
		assert.Empty(warnings)
		assert.Len(errors, 1)
	}
}

func TestValidationUUIDWithLocalityWithValidUUIDReturnNothing(t *testing.T) {
	assert := assert.New(t)

	for _, uuid := range []string{"fr-par/6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		warnings, errors := verify.IsUUIDWithLocality()(uuid, "key")
		assert.Empty(warnings)
		assert.Empty(errors)
	}
}

func TestValidationUUIDWithLocalityWithInvalidUUIDReturnError(t *testing.T) {
	assert := assert.New(t)

	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		warnings, errors := verify.IsUUIDWithLocality()(uuid, "key")
		assert.Empty(warnings)
		assert.Len(errors, 1, uuid)
	}
}
