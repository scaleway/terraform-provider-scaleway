package verify_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
	"github.com/stretchr/testify/assert"
)

func TestValidationUUIDWithInvalidUUIDReturnError(t *testing.T) {
	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid"} {
		diags := verify.IsUUID()(uuid, cty.Path{})
		assert.Len(t, diags, 1)
	}
}

func TestValidationUUIDWithValidUUIDReturnNothing(t *testing.T) {
	diags := verify.IsUUID()("6ba7b810-9dad-11d1-80b4-00c04fd430c8", cty.Path{})

	assert.Empty(t, diags)
}

func TestValidationUUIDorUUIDWithLocalityWithValidUUIDReturnNothing(t *testing.T) {
	for _, uuid := range []string{"fr-par/6ba7b810-9dad-11d1-80b4-00c04fd430c8", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		diags := verify.IsUUIDorUUIDWithLocality()(uuid, cty.Path{})
		assert.Empty(t, diags)
	}
}

func TestValidationUUIDorUUIDWithLocalityWithInvalidUUIDReturnError(t *testing.T) {
	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid"} {
		diags := verify.IsUUIDorUUIDWithLocality()(uuid, cty.Path{})
		assert.Len(t, diags, 1)
	}
}

func TestValidationUUIDWithLocalityWithValidUUIDReturnNothing(t *testing.T) {
	for _, uuid := range []string{"fr-par/6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		diags := verify.IsUUIDWithLocality()(uuid, cty.Path{})
		assert.Empty(t, diags)
	}
}

func TestValidationUUIDWithLocalityWithInvalidUUIDReturnError(t *testing.T) {
	for _, uuid := range []string{"fr-par/wrong-uuid/resource", "fr-par/wrong-uuid", "wrong-uuid", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"} {
		diags := verify.IsUUIDWithLocality()(uuid, cty.Path{})
		assert.Len(t, diags, 1, uuid)
	}
}
