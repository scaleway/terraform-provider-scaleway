package verify

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// IsUUIDorUUIDWithLocality validates the schema is a UUID or the combination of a locality and a UUID
// e.g. "6ba7b810-9dad-11d1-80b4-00c04fd430c8" or "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func IsUUIDorUUIDWithLocality() schema.SchemaValidateDiagFunc {
	return func(value interface{}, path cty.Path) diag.Diagnostics {
		return IsUUID()(locality.ExpandID(value), path)
	}
}

// IsUUID validates the schema following the canonical UUID format
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func IsUUID() schema.SchemaValidateDiagFunc {
	return func(value interface{}, path cty.Path) diag.Diagnostics {
		uuid, isString := value.(string)
		if !isString {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "invalid UUID not a string",
				AttributePath: path,
			}}
		}

		if !validation.IsUUID(uuid) {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "invalid UUID: " + uuid,
				AttributePath: path,
				Detail:        "format should be 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' (36) and contains valid hexadecimal characters",
			}}
		}

		return nil
	}
}

func IsUUIDWithLocality() schema.SchemaValidateDiagFunc {
	return func(value interface{}, path cty.Path) diag.Diagnostics {
		uuid, isString := value.(string)
		if !isString {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "invalid UUID not a string",
				AttributePath: path,
			}}
		}

		_, subUUID, err := locality.ParseLocalizedID(uuid)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "invalid UUID with locality: " + uuid,
				AttributePath: path,
				Detail:        "format should be 'locality/uuid'",
			}}
		}

		return IsUUID()(subUUID, path)
	}
}
