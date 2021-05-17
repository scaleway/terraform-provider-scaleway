package scaleway

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/validation"
)

// validationUUID validates the schema is a UUID or the combination of a locality and a UUID
// e.g. "6ba7b810-9dad-11d1-80b4-00c04fd430c8" or "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func validationUUIDorUUIDWithLocality() func(interface{}, string) ([]string, []error) {
	return func(v interface{}, key string) ([]string, []error) {
		return validationUUID()(expandID(v), key)
	}
}

// validationUUID validates the schema following the canonical UUID format
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func validationUUID() func(interface{}, string) ([]string, []error) {
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
