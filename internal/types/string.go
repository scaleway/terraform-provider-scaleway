package types

import (
	"reflect"
	"slices"
	"sort"

	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func FlattenStringPtr(s *string) any {
	if s == nil {
		return ""
	}

	return *s
}

func ExpandStringPtr(data any) *string {
	if data == nil || data == "" {
		return nil
	}

	return new(data.(string))
}

// NewRandomName returns a random name prefixed for terraform.
func NewRandomName(prefix string) string {
	return namegenerator.GetRandomName("tf", prefix)
}

func ExpandOrGenerateString(data any, prefix string) string {
	if data == nil || data == "" {
		return NewRandomName(prefix)
	}

	return data.(string)
}

func ExpandStringWithDefault(data any, defaultValue string) string {
	if data == nil || data.(string) == "" {
		return defaultValue
	}

	return data.(string)
}

func ExpandSliceStringPtr(data any) []*string {
	if data == nil {
		return nil
	}

	stringSlice := []*string(nil)
	for _, s := range data.([]any) {
		stringSlice = append(stringSlice, ExpandStringPtr(s))
	}

	return stringSlice
}

func FlattenSliceStringPtr(s []*string) any {
	res := make([]any, 0, len(s))
	for _, strPtr := range s {
		res = append(res, FlattenStringPtr(strPtr))
	}

	return res
}

func FlattenSliceString(s []string) any {
	res := make([]any, 0, len(s))
	for _, strPtr := range s {
		res = append(res, strPtr)
	}

	return res
}

func ExpandUpdatedStringPtr(data any) *string {
	str := ""
	if data != nil {
		str = data.(string)
	}

	return &str
}

// expandStringSlice converts a schema []any of strings into []string.
// Nil data or a non-slice yields an empty slice.
func expandStringSlice(data any) []string {
	raw, ok := data.([]any)
	if !ok || data == nil {
		return []string{}
	}

	stringSlice := make([]string, 0, len(raw))
	for _, s := range raw {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}

		stringSlice = append(stringSlice, s.(string))
	}

	return stringSlice
}

func ExpandStrings(data any) []string {
	return expandStringSlice(data)
}

func ExpandStringsPtr(data any) *[]string {
	if _, ok := data.([]any); !ok || data == nil {
		return nil
	}

	stringSlice := expandStringSlice(data)
	if len(stringSlice) == 0 {
		return nil
	}

	return &stringSlice
}

// ExpandUpdatedStringsPtr expands a string slice but will default to an empty list.
// Should be used on schema update so emptying a list will update resource.
func ExpandUpdatedStringsPtr(data any) *[]string {
	stringSlice := expandStringSlice(data)

	return &stringSlice
}

func ExpandSliceIDsPtr(rawIDs any) *[]string {
	ids := locality.ExpandIDs(rawIDs)

	return &ids
}

func FlattenSliceIDs(certificates []string, zone scw.Zone) any {
	res := make([]any, 0, len(certificates))
	for _, certificateID := range certificates {
		res = append(res, zonal.NewIDString(zone, certificateID))
	}

	return res
}

func SliceContainsString(slice []string, str string) bool {
	return slices.Contains(slice, str)
}

func CompareStringListsIgnoringOrder(oldListStr, newListStr []string) bool {
	if len(oldListStr) != len(newListStr) {
		return false // different lengths means there's definitely a change
	}

	sort.Strings(oldListStr)
	sort.Strings(newListStr)

	return reflect.DeepEqual(oldListStr, newListStr)
}
