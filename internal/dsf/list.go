package dsf

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// OrderDiff suppresses diffs for TypeList attributes when the only change is the order of elements.
// https://github.com/hashicorp/terraform-plugin-sdk/issues/477#issuecomment-1238807249
func OrderDiff(k, _, _ string, d *schema.ResourceData) bool {
	baseKey := ExtractBaseKey(k)
	oldList, newList := GetStringListsFromState(baseKey, d)

	return types.CompareStringListsIgnoringOrder(oldList, newList)
}

func ExtractBaseKey(k string) string {
	lastDotIndex := strings.LastIndex(k, ".")
	if lastDotIndex != -1 {
		return k[:lastDotIndex]
	}

	return k
}

func GetStringListsFromState(key string, d *schema.ResourceData) ([]string, []string) {
	oldList, newList := d.GetChange(key)

	oldListStr := make([]string, len(oldList.([]any)))
	newListStr := make([]string, len(newList.([]any)))

	for i, v := range oldList.([]any) {
		oldListStr[i] = fmt.Sprint(v)
	}

	for i, v := range newList.([]any) {
		newListStr[i] = fmt.Sprint(v)
	}

	return oldListStr, newListStr
}
