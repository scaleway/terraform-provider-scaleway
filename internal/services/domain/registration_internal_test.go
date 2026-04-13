package domain

import "testing"

func TestRegistrationOwnerContactIDValidationAcceptsOpaqueID(t *testing.T) {
	t.Parallel()

	ownerContactIDSchema, exists := registrationSchema()["owner_contact_id"]
	if !exists {
		t.Fatal("owner_contact_id schema not found")
	}

	opaqueContactID := "eV_ExampleContactId0123456789abcdefghijklmnopqrstu="

	if ownerContactIDSchema.ValidateFunc == nil {
		return
	}

	if _, errors := ownerContactIDSchema.ValidateFunc(opaqueContactID, "owner_contact_id"); len(errors) > 0 {
		t.Fatalf("expected opaque owner contact ID to be accepted, got validation errors: %v", errors)
	}
}
