package cdf

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// expandListKeys return the list of keys for an attribute in a list
// example for private-networks.#.id in a list of size 2
// will return private-networks.0.id and private-networks.1.id
// additional_volume_ids.#
// will return additional_volume_ids.0 and additional_volume_ids.1
func expandListKeys(key string, diff *schema.ResourceDiff) []string {
	addr := strings.Split(key, ".")
	// index of # in the addr
	index := 0

	for i := range addr {
		if addr[i] == "#" {
			index = i
		}
	}

	// get attribute.#
	listKey := key[:strings.Index(key, "#")+1]
	listLength := diff.Get(listKey).(int)

	keys := make([]string, 0, listLength)

	for i := range listLength {
		addr[index] = strconv.FormatInt(int64(i), 10)
		keys = append(keys, strings.Join(addr, "."))
	}

	return keys
}

// getLocality find the locality of a resource
// Will try to get the zone if available then use region
// Will also use default zone or region if available
func getLocality(diff *schema.ResourceDiff, m any) string {
	var loc string

	rawStateType := diff.GetRawState().Type()

	if rawStateType.HasAttribute("zone") {
		zone, _ := meta.ExtractZone(diff, m)
		loc = zone.String()
	} else if rawStateType.HasAttribute("region") {
		region, _ := meta.ExtractRegion(diff, m)
		loc = region.String()
	}

	return loc
}

// LocalityCheck create a function that will validate locality IDs stored in given keys
// This locality IDs should have the same locality as the resource
// It will search for zone or region in resource.
// Should not be used on computed keys, if a computed key is going to change on zone/region change
// this function will still block the terraform plan
func LocalityCheck(keys ...string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, m any) error {
		l := getLocality(diff, m)

		if l == "" {
			return errors.New("missing locality zone or region to check IDs")
		}

		for _, key := range keys {
			// Handle values in lists
			if strings.Contains(key, "#") {
				listKeys := expandListKeys(key, diff)

				for _, listKey := range listKeys {
					IDLocality, _, err := locality.ParseLocalizedID(diff.Get(listKey).(string))
					if err == nil && !locality.CompareLocalities(IDLocality, l) {
						return fmt.Errorf("given %s %s has different locality than the resource %q", listKey, diff.Get(listKey), l)
					}
				}
			} else {
				IDLocality, _, err := locality.ParseLocalizedID(diff.Get(key).(string))
				if err == nil && !locality.CompareLocalities(IDLocality, l) {
					return fmt.Errorf("given %s %s has different locality than the resource %q", key, diff.Get(key), l)
				}
			}
		}

		return nil
	}
}
