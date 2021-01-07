package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// validationUUID validates the schema is a UUID or the combination of a locality and a UUID
// e.g. "6ba7b810-9dad-11d1-80b4-00c04fd430c8" or "fr-par/6ba7b810-9dad-11d1-80b4-00c04fd430c8".
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

// validationZone validates the schema is a valid zone
func validationZone() func(interface{}, string) ([]string, []error) {
	return func(v interface{}, key string) (warnings []string, errors []error) {
		rawZone, isString := v.(string)
		if !isString {
			return nil, []error{fmt.Errorf("invalid zone: not a string")}
		}

		// TODO: Use scw.ParseZone when the format validation will be implemented.
		zone, _ := scw.ParseZone(rawZone)
		if rawZone == "par1" || rawZone == "ams1" {
			warnings = append(warnings, fmt.Sprintf("%s is a deprecated name for zone, use %v instead", rawZone, zone))
		} else if !zone.Exists() {
			warnings = append(warnings, fmt.Sprintf("%s zone is not recognized", rawZone))
		}

		return
	}
}

// validationRegion validates the schema is a valid region
func validationRegion() func(interface{}, string) ([]string, []error) {
	return func(v interface{}, key string) (warnings []string, errors []error) {
		rawRegion, isString := v.(string)
		if !isString {
			return nil, []error{fmt.Errorf("invalid region: not a string")}
		}

		// TODO: Use scw.ParseRegion when the format validation will be implemented.
		region, _ := scw.ParseRegion(rawRegion)
		if rawRegion == "par1" || rawRegion == "ams1" {
			warnings = append(warnings, fmt.Sprintf("%s is a deprecated name for region, use %v instead", rawRegion, region))
		} else if !region.Exists() {
			warnings = append(warnings, fmt.Sprintf("%s region is not recognized", rawRegion))
		}

		return
	}
}

// validationStringNotInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type string and does not match the value of an element in the invalid slice
// will test with in lower case if ignoreCase is true
func validationStringNotInSlice(invalid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		for _, str := range invalid {
			if v == str || (ignoreCase && strings.EqualFold(v, str)) {
				es = append(es, fmt.Errorf("expected %s not to be one of %v, got %s", k, invalid, v))
			}
		}

		return
	}
}
