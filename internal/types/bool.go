package types

import "github.com/scaleway/scaleway-sdk-go/scw"

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
