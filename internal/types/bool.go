package types

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func FlattenBoolPtr(b *bool) interface{} {
	if b == nil {
		return nil
	}
	return *b
}

func ExpandBoolPtr(data interface{}) *bool {
	if data == nil {
		return nil
	}
	return scw.BoolPtr(data.(bool))
}

func GetBool(d *schema.ResourceData, key string) interface{} {
	val, ok := d.GetOkExists(key) //nolint:staticcheck
	if !ok {
		return nil
	}
	return val
}
