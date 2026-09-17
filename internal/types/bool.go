package types

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func FlattenBoolPtr(b *bool) any {
	if b == nil {
		return nil
	}

	return *b
}

func ExpandBoolPtr(data any) *bool {
	if data == nil {
		return nil
	}

	return new(data.(bool))
}

// rawConfigReader abstracts ResourceData and ResourceDiff for GetBool.
type rawConfigReader interface {
	GetRawConfig() cty.Value
}

var (
	_ rawConfigReader = (*schema.ResourceData)(nil)
	_ rawConfigReader = (*schema.ResourceDiff)(nil)
)

// GetBool returns the configured boolean value, or nil when the attribute is absent.
//
// It reads the raw configuration because GetOk treats false as unset, which would
// prevent callers from distinguishing an explicit false from an omitted value.
func GetBool(d rawConfigReader, key string) any {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return nil
	}

	value, exists := rawConfig.AsValueMap()[key]
	if !exists || value.IsNull() || !value.IsKnown() {
		return nil
	}

	return value.True()
}
