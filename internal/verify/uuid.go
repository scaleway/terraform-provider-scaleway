package verify

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// IsUUIDorUUIDWithLocality validates the schema is a UUID or the combination of a locality and a UUID
// e.g. "6ba7b810-9dad-11d1-80b4-00c04fd430c8" or "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func IsUUIDorUUIDWithLocality() schema.SchemaValidateFunc {
	return func(v interface{}, key string) ([]string, []error) {
		return IsUUID()(locality.ExpandID(v), key)
	}
}

// IsUUID validates the schema following the canonical UUID format
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func IsUUID() schema.SchemaValidateFunc {
	return func(v interface{}, key string) (warnings []string, errors []error) {
		uuid, isString := v.(string)
		if !isString {
			return nil, []error{fmt.Errorf("invalid UUID for key '%s': not a string", key)}
		}

		if !validation.IsUUID(uuid) {
			return nil, []error{fmt.Errorf("invalid UUID for key '%s': '%s' (%d): format should be 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' (36) and contains valid hexadecimal characters", key, uuid, len(uuid))}
		}

		return
	}
}

func IsUUIDWithLocality() schema.SchemaValidateFunc {
	return func(v interface{}, key string) (warnings []string, errors []error) {
		uuid, isString := v.(string)
		if !isString {
			errors = []error{fmt.Errorf("invalid UUID for key '%s': not a string", key)}
			return
		}
		_, subUUID, err := locality.ParseLocalizedID(uuid)
		if err != nil {
			errors = []error{fmt.Errorf("invalid UUID with locality for key  '%s': '%s' (%d): format should be 'locality/uuid'", key, uuid, len(uuid))}
			return
		}
		return IsUUID()(subUUID, key)
	}
}
