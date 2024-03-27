package types

import (
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func FlattenStringPtr(s *string) interface{} {
	if s == nil {
		return ""
	}
	return *s
}

func ExpandStringPtr(data interface{}) *string {
	if data == nil || data == "" {
		return nil
	}
	return scw.StringPtr(data.(string))
}

// NewRandomName returns a random name prefixed for terraform.
func NewRandomName(prefix string) string {
	return namegenerator.GetRandomName("tf", prefix)
}

func ExpandOrGenerateString(data interface{}, prefix string) string {
	if data == nil || data == "" {
		return NewRandomName(prefix)
	}
	return data.(string)
}

func ExpandStringWithDefault(data interface{}, defaultValue string) string {
	if data == nil || data.(string) == "" {
		return defaultValue
	}
	return data.(string)
}

func ExpandSliceStringPtr(data interface{}) []*string {
	if data == nil {
		return nil
	}
	stringSlice := []*string(nil)
	for _, s := range data.([]interface{}) {
		stringSlice = append(stringSlice, ExpandStringPtr(s))
	}
	return stringSlice
}

func FlattenSliceStringPtr(s []*string) interface{} {
	res := make([]interface{}, 0, len(s))
	for _, strPtr := range s {
		res = append(res, FlattenStringPtr(strPtr))
	}
	return res
}

func FlattenSliceString(s []string) interface{} {
	res := make([]interface{}, 0, len(s))
	for _, strPtr := range s {
		res = append(res, strPtr)
	}
	return res
}

func ExpandUpdatedStringPtr(data interface{}) *string {
	str := ""
	if data != nil {
		str = data.(string)
	}
	return &str
}

func ExpandStrings(data interface{}) []string {
	stringSlice := make([]string, 0, len(data.([]interface{})))
	for _, s := range data.([]interface{}) {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}
		stringSlice = append(stringSlice, s.(string))
	}
	return stringSlice
}

func ExpandStringsPtr(data interface{}) *[]string {
	stringSlice := make([]string, 0, len(data.([]interface{})))
	if _, ok := data.([]interface{}); !ok || data == nil {
		return nil
	}
	for _, s := range data.([]interface{}) {
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
func ExpandUpdatedStringsPtr(data interface{}) *[]string {
	stringSlice := []string{}
	if _, ok := data.([]interface{}); !ok || data == nil {
		return &stringSlice
	}
	for _, s := range data.([]interface{}) {
		// zero-value is nil, ["foo", ""]
		if s == nil {
			s = ""
		}
		stringSlice = append(stringSlice, s.(string))
	}
	return &stringSlice
}

func ExpandSliceIDsPtr(rawIDs interface{}) *[]string {
	stringSlice := make([]string, 0, len(rawIDs.([]interface{})))
	if _, ok := rawIDs.([]interface{}); !ok || rawIDs == nil {
		return &stringSlice
	}
	for _, s := range rawIDs.([]interface{}) {
		stringSlice = append(stringSlice, locality.ExpandID(s.(string)))
	}
	return &stringSlice
}

func ExpandStringsOrEmpty(data interface{}) []string {
	stringSlice := make([]string, 0, len(data.([]interface{})))
	if _, ok := data.([]interface{}); !ok || data == nil {
		return stringSlice
	}
	for _, s := range data.([]interface{}) {
		stringSlice = append(stringSlice, s.(string))
	}
	return stringSlice
}

func FlattenSliceIDs(certificates []string, zone scw.Zone) interface{} {
	res := []interface{}(nil)
	for _, certificateID := range certificates {
		res = append(res, zonal.NewIDString(zone, certificateID))
	}

	return res
}
