package dsf

import (
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IgnoreCase(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.EqualFold(oldValue, newValue)
}

func IgnoreCaseAndHyphen(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.ReplaceAll(strings.ToLower(oldValue), "-", "_") == strings.ReplaceAll(strings.ToLower(newValue), "-", "_")
}

func CompareArgon2idPasswordAndHash(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	match, err := argon2id.ComparePasswordAndHash(newValue, oldValue)
	if err != nil {
		return false
	}

	return match
}
