package domain_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression for #4554: flattenContactExtensionNL returns a list of maps; the
// contact schema must accept that shape on d.Set or owner/admin/tech contacts
// are silently nulled.
func TestRegistrationContactSchema_ExtensionNL(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, domain.ResourceRegistration().SchemaFunc(), map[string]any{
		"domain_names":     []any{"example.nl"},
		"owner_contact_id": "11111111-1111-1111-1111-111111111111",
	})

	err := d.Set("owner_contact", []map[string]any{{
		"phone_number":   "0612345678",
		"legal_form":     "corporate",
		"firstname":      "Jane",
		"lastname":       "Doe",
		"email":          "jane@example.com",
		"address_line_1": "1 Main",
		"zip":            "1000AA",
		"city":           "Amsterdam",
		"country":        "NL",
		"extension_nl": []map[string]any{{
			"legal_form":                     "non_dutch_eu_company",
			"legal_form_registration_number": "12345678",
		}},
	}})
	require.NoError(t, err)
	assert.Equal(t, "non_dutch_eu_company", d.Get("owner_contact.0.extension_nl.0.legal_form"))
	assert.Equal(t, "12345678", d.Get("owner_contact.0.extension_nl.0.legal_form_registration_number"))
}
