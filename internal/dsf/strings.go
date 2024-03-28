package dsf

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IgnoreCase(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.EqualFold(oldValue, newValue)
}

func IgnoreCaseAndHyphen(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.ReplaceAll(strings.ToLower(oldValue), "-", "_") == strings.ReplaceAll(strings.ToLower(newValue), "-", "_")
}
