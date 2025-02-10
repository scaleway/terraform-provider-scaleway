package types

import "github.com/scaleway/scaleway-sdk-go/scw"

func FlattenInt32Ptr(i *int32) interface{} {
	if i == nil {
		return 0
	}

	return *i
}

func FlattenUint32Ptr(i *uint32) interface{} {
	if i == nil {
		return 0
	}

	return *i
}

func ExpandInt32Ptr(data interface{}) *int32 {
	if data == nil || data == "" {
		return nil
	}

	return scw.Int32Ptr(int32(data.(int)))
}

func ExpandUint32Ptr(data interface{}) *uint32 {
	if data == nil || data == "" {
		return nil
	}

	return scw.Uint32Ptr(uint32(data.(int)))
}
