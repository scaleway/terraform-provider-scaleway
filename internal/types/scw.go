package types

import "github.com/scaleway/scaleway-sdk-go/scw"

func FlattenSize(size *scw.Size) any {
	if size == nil {
		return 0
	}

	return *size
}

func ExpandSize(data any) *scw.Size {
	if data == nil || data == "" {
		return nil
	}

	size := scw.Size(data.(int))

	return &size
}
