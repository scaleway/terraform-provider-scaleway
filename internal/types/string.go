package types

import (
	"hash/crc32"
	"reflect"
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

	return scw.StringPtr(data.(string))
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

func ExpandStrings(data any) []string {
	stringSlice := make([]string, 0, len(data.([]any)))

	for _, s := range data.([]any) {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}

		stringSlice = append(stringSlice, s.(string))
	}

	return stringSlice
}

func ExpandStringsPtr(data any) *[]string {
	stringSlice := make([]string, 0, len(data.([]any)))

	if _, ok := data.([]any); !ok || data == nil {
		return nil
	}

	for _, s := range data.([]any) {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}

		stringSlice = append(stringSlice, s.(string))
	}

	if len(stringSlice) == 0 {
		return nil
	}

	return &stringSlice
}

// ExpandUpdatedStringsPtr expands a string slice but will default to an empty list.
// Should be used on schema update so emptying a list will update resource.
func ExpandUpdatedStringsPtr(data any) *[]string {
	stringSlice := []string{}
	if _, ok := data.([]any); !ok || data == nil {
		return &stringSlice
	}

	for _, s := range data.([]any) {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}

		stringSlice = append(stringSlice, s.(string))
	}

	return &stringSlice
}

func ExpandSliceIDs(rawIDs any) []string {
	stringSlice := make([]string, 0, len(rawIDs.([]any)))
	if _, ok := rawIDs.([]any); !ok || rawIDs == nil {
		return stringSlice
	}

	for _, s := range rawIDs.([]any) {
		stringSlice = append(stringSlice, locality.ExpandID(s.(string)))
	}

	return stringSlice
}

func ExpandSliceIDsPtr(rawIDs any) *[]string {
	stringSlice := make([]string, 0, len(rawIDs.([]any)))
	if _, ok := rawIDs.([]any); !ok || rawIDs == nil {
		return &stringSlice
	}

	for _, s := range rawIDs.([]any) {
		stringSlice = append(stringSlice, locality.ExpandID(s.(string)))
	}

	return &stringSlice
}

func ExpandStringsOrEmpty(data any) []string {
	stringSlice := make([]string, 0, len(data.([]any)))
	if _, ok := data.([]any); !ok || data == nil {
		return stringSlice
	}

	for _, s := range data.([]any) {
		stringSlice = append(stringSlice, s.(string))
	}

	return stringSlice
}

func FlattenSliceIDs(certificates []string, zone scw.Zone) any {
	res := make([]any, 0, len(certificates))
	for _, certificateID := range certificates {
		res = append(res, zonal.NewIDString(zone, certificateID))
	}

	return res
}

// StringHashcode hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non-negative integer. Here we cast to an integer
// and invert it if the result is negative.
func StringHashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}

	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func SliceContainsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}

	return false
}

func CompareStringListsIgnoringOrder(oldListStr, newListStr []string) bool {
	if len(oldListStr) != len(newListStr) {
		return false // different lengths means there's definitely a change
	}

	sort.Strings(oldListStr)
	sort.Strings(newListStr)

	return reflect.DeepEqual(oldListStr, newListStr)
}
