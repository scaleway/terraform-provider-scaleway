package lb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	validator "github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultLbLbTimeout = 10 * time.Minute
	RetryLbIPInterval  = 5 * time.Second
)

// lbAPIWithZone returns an lb API WITH zone for a Create request
func lbAPIWithZone(d *schema.ResourceData, m interface{}) (*lbSDK.ZonedAPI, scw.Zone, error) {
	lbAPI := lbSDK.NewZonedAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return lbAPI, zone, nil
}

// NewAPIWithZoneAndID returns an lb API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m interface{}, id string) (*lbSDK.ZonedAPI, scw.Zone, string, error) {
	lbAPI := lbSDK.NewZonedAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return lbAPI, zone, ID, nil
}

func IsPrivateNetworkEqual(a, b interface{}) bool {
	// Find out the diff Private Network or not
	if _, ok := a.(*lbSDK.PrivateNetwork); ok {
		if _, ok := b.(*lbSDK.PrivateNetwork); ok {
			if a.(*lbSDK.PrivateNetwork).PrivateNetworkID == b.(*lbSDK.PrivateNetwork).PrivateNetworkID {
				// if both has dhcp config should not update
				if a.(*lbSDK.PrivateNetwork).DHCPConfig != nil && b.(*lbSDK.PrivateNetwork).DHCPConfig != nil {
					return true
				}
				// check static config
				aConfig := a.(*lbSDK.PrivateNetwork).StaticConfig
				bConfig := b.(*lbSDK.PrivateNetwork).StaticConfig
				if aConfig != nil && bConfig != nil {
					// check if static config is different
					return reflect.DeepEqual(aConfig.IPAddress, bConfig.IPAddress)
				}
			}
		}
	}
	return false
}

func privateNetworksCompare(slice1, slice2 []*lbSDK.PrivateNetwork) []*lbSDK.PrivateNetwork {
	var diff []*lbSDK.PrivateNetwork

	m := make(map[string]struct{}, len(slice1))
	for _, pn := range slice1 {
		m[pn.PrivateNetworkID] = struct{}{}
	}
	// find the differences
	for _, pn := range slice2 {
		if _, foundID := m[pn.PrivateNetworkID]; !foundID || (foundID && !IsPrivateNetworkEqual(slice1, slice2)) {
			diff = append(diff, pn)
		}
	}
	return diff
}

func lbUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

// lbUpgradeV1UpgradeFunc allow upgrade the from regional to a zoned resource.
func UpgradeStateV1Func(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	var err error
	// element id: upgrade
	ID, exist := rawState["id"]
	if !exist {
		return nil, errors.New("upgrade: id not exist")
	}
	rawState["id"], err = lbUpgradeV1RegionalToZonedID(ID.(string))
	if err != nil {
		return nil, err
	}
	// return rawState updated
	return rawState, nil
}

func lbUpgradeV1RegionalToZonedID(element string) (string, error) {
	loc, id, err := locality.ParseLocalizedID(element)
	// return error if l cannot be parsed
	if err != nil {
		return "", fmt.Errorf("upgrade: could not retrieve the locality from `%s`", element)
	}
	// if locality is already zoned return
	if validator.IsZone(loc) {
		return element, nil
	}
	//  append zone 1 as default: e.g. fr-par-1
	return fmt.Sprintf("%s-1/%s", loc, id), nil
}

func ipv4Match(cidr, ipStr string) bool {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(ipStr)

	return cidrNet.Contains(ip)
}

func lbPrivateNetworkSetHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if pnID, ok := m["private_network_id"]; ok {
		buf.WriteString(locality.ExpandID(pnID))
	}

	if staticConfig, ok := m["static_config"]; ok && len(staticConfig.([]interface{})) > 0 {
		for _, item := range staticConfig.([]interface{}) {
			buf.WriteString(item.(string))
		}
	}

	return types.StringHashcode(buf.String())
}

func diffSuppressFunc32SubnetMask(k, _, _ string, d *schema.ResourceData) bool {
	baseKey := dsf.ExtractBaseKey(k)
	oldList, newList := dsf.GetStringListsFromState(baseKey, d)

	oldList = normalizeIPSubnetList(oldList)
	newList = normalizeIPSubnetList(newList)

	return types.CompareStringListsIgnoringOrder(oldList, newList)
}

func normalizeIPSubnetList(list []string) []string {
	normalized := make([]string, len(list))
	for i, ip := range list {
		normalized[i] = normalizeIPSubnet(ip)
	}

	return normalized
}

func normalizeIPSubnet(ip string) string {
	if strings.HasSuffix(ip, "/32") {
		return strings.TrimSuffix(ip, "/32")
	}
	return ip
}

// getRawConfigForKey returns the value for a specific key in the user's raw configuration, which can be useful on resources' update
// The value for the key to look for must be a primitive type (bool, string, number) and the expected type of the value should be passed as the ty parameter
func getRawConfigForKey(d *schema.ResourceData, key string, ty cty.Type) (interface{}, bool) {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return nil, false
	}
	return GetKeyInRawConfigMap(rawConfig.AsValueMap(), key, ty)
}

func GetKeyInRawConfigMap(rawConfig map[string]cty.Value, key string, ty cty.Type) (interface{}, bool) {
	if key == "" {
		return rawConfig, false
	}
	// We split the key into its elements
	keys := strings.Split(key, ".")

	// We look at the first element's type
	if value, ok := rawConfig[keys[0]]; ok {
		switch {
		case value.Type().IsListType():
			// If it's a list and the second element of the key is an index, we look for the value in the list at the given index
			if index, err := strconv.Atoi(keys[1]); err == nil {
				return GetKeyInRawConfigMap(value.AsValueSlice()[index].AsValueMap(), strings.Join(keys[2:], ""), ty)
			}
			// If it's a list and the second element of the key is '#', we look for the value in the list's first element
			return GetKeyInRawConfigMap(value.AsValueSlice()[0].AsValueMap(), strings.Join(keys[2:], ""), ty)

		case value.Type().IsMapType():
			// If it's a map, we look for the value in the map
			return GetKeyInRawConfigMap(value.AsValueMap(), strings.Join(keys[1:], ""), ty)

		case value.Type().IsPrimitiveType():
			// If it's a primitive type (bool, string, number), we convert the value to the expected type given as parameter before returning it
			switch ty {
			case cty.String:
				if value.IsNull() {
					return nil, false
				}
				return value.AsString(), true
			case cty.Bool:
				if value.IsNull() {
					return false, false
				}
				if value.True() {
					return true, true
				}
				return false, true
			case cty.Number:
				if value.IsNull() {
					return nil, false
				}
				valueInt, _ := value.AsBigFloat().Int64()
				return valueInt, true
			}
		}
	}
	return nil, false
}

func customizeDiffLBIPIDs(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	oldIPIDs, newIPIDs := diff.GetChange("ip_ids")
	oldIPIDsSet := make(map[string]struct{})
	newIPIDsSet := make(map[string]struct{})

	for _, id := range oldIPIDs.([]interface{}) {
		oldIPIDsSet[id.(string)] = struct{}{}
	}

	for _, id := range newIPIDs.([]interface{}) {
		newIPIDsSet[id.(string)] = struct{}{}
	}

	// Check if any IP ID is being removed
	for id := range oldIPIDsSet {
		if _, ok := newIPIDsSet[id]; !ok {
			return diff.ForceNew("ip_ids")
		}
	}

	return nil
}

func customizeDiffAssignFlexibleIPv6(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	oldValue, newValue := diff.GetChange("assign_flexible_ipv6")
	if oldValue.(bool) && !newValue.(bool) {
		return diff.ForceNew("assign_flexible_ipv6")
	}
	return nil
}
