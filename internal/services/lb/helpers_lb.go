package lb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
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
	defaultLbLbTimeout = 15 * time.Minute
	RetryLbIPInterval  = 5 * time.Second
)

// lbAPIWithZone returns an lb API WITH zone for a Create request
func lbAPIWithZone(d *schema.ResourceData, m any) (*lbSDK.ZonedAPI, scw.Zone, error) {
	lbAPI := lbSDK.NewZonedAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return lbAPI, zone, nil
}

// NewAPIWithZoneAndID returns an lb API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, id string) (*lbSDK.ZonedAPI, scw.Zone, string, error) {
	lbAPI := lbSDK.NewZonedAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return lbAPI, zone, ID, nil
}

func IsPrivateNetworkEqual(a, b *lbSDK.PrivateNetwork) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.PrivateNetworkID != b.PrivateNetworkID {
		return false
	}

	if !reflect.DeepEqual(a.DHCPConfig, b.DHCPConfig) {
		return false
	}

	if !reflect.DeepEqual(a.StaticConfig, b.StaticConfig) {
		return false
	}

	return true
}

func PrivateNetworksCompare(oldPNs, newPNs []*lbSDK.PrivateNetwork) ([]*lbSDK.PrivateNetwork, []*lbSDK.PrivateNetwork) {
	var toDetach, toAttach []*lbSDK.PrivateNetwork

	oldPNMap := make(map[string]*lbSDK.PrivateNetwork, len(oldPNs))
	for _, pn := range oldPNs {
		oldPNMap[pn.PrivateNetworkID] = pn
	}

	newPNMap := make(map[string]*lbSDK.PrivateNetwork, len(newPNs))
	for _, pn := range newPNs {
		newPNMap[pn.PrivateNetworkID] = pn
	}

	for id, oldPN := range oldPNMap {
		newPN, found := newPNMap[id]
		if !found {
			toDetach = append(toDetach, oldPN)
		} else if !IsPrivateNetworkEqual(oldPN, newPN) {
			toDetach = append(toDetach, oldPN)
			toAttach = append(toAttach, newPN)
		}
	}

	for id, newPN := range newPNMap {
		if _, found := oldPNMap[id]; !found {
			toAttach = append(toAttach, newPN)
		}
	}

	return toDetach, toAttach
}

func lbUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

// lbUpgradeV1UpgradeFunc allow upgrade the from regional to a zoned resource.
func UpgradeStateV1Func(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
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

func lbPrivateNetworkSetHash(v any) int {
	var buf bytes.Buffer

	m := v.(map[string]any)
	if pnID, ok := m["private_network_id"]; ok {
		buf.WriteString(locality.ExpandID(pnID))
	}

	if staticConfig, ok := m["static_config"]; ok && len(staticConfig.([]any)) > 0 {
		for _, item := range staticConfig.([]any) {
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
	if before, ok := strings.CutSuffix(ip, "/32"); ok {
		return before
	}

	return ip
}

func customizeDiffLBIPIDs(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	oldIPIDs, newIPIDs := diff.GetChange("ip_ids")
	oldIPIDsSet := make(map[string]struct{})
	newIPIDsSet := make(map[string]struct{})

	for _, id := range oldIPIDs.([]any) {
		oldIPIDsSet[id.(string)] = struct{}{}
	}

	for _, id := range newIPIDs.([]any) {
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

func customizeDiffAssignFlexibleIPv6(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	oldValue, newValue := diff.GetChange("assign_flexible_ipv6")
	if oldValue.(bool) && !newValue.(bool) {
		return diff.ForceNew("assign_flexible_ipv6")
	}

	return nil
}

func ResourceLBPrivateNetworkParseID(resourceID string) (zone scw.Zone, lbID string, pnID string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Zone(idParts[0]), idParts[1], idParts[2], nil
}
