package types

import "github.com/scaleway/scaleway-sdk-go/scw"

func FlattenSize(size *scw.Size) interface{} {
	if size == nil {
		return 0
	}
	return *size
}
