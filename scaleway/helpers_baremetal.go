package scaleway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	baremetalServerWaitForTimeout   = 60 * time.Minute
	baremetalServerRetryFuncTimeout = baremetalServerWaitForTimeout + time.Minute // some RetryFunc are calling a WaitFor
	defaultBaremetalServerTimeout   = baremetalServerRetryFuncTimeout + time.Minute
	baremetalRetryInterval          = 5 * time.Second
)

// instanceAPIWithZone returns a new baremetal API and the zone for a Create request
func baremetalAPIWithZone(d *schema.ResourceData, m interface{}) (*baremetal.API, scw.Zone, error) {
	baremetalAPI := baremetal.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return baremetalAPI, zone, nil
}

// instanceAPIWithZoneAndID returns an baremetal API with zone and ID extracted from the state
func BaremetalAPIWithZoneAndID(m interface{}, id string) (*baremetal.API, zonal.ID, error) {
	baremetalAPI := baremetal.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, zonal.ID{}, err
	}
	return baremetalAPI, zonal.NewID(zone, ID), nil
}

// returns a new baremetal private network API and the zone for a Create request
func baremetalPrivateNetworkAPIWithZone(d *schema.ResourceData, m interface{}) (*baremetal.PrivateNetworkAPI, scw.Zone, error) {
	baremetalPrivateNetworkAPI := baremetal.NewPrivateNetworkAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return baremetalPrivateNetworkAPI, zone, nil
}

// BaremetalPrivateNetworkAPIWithZoneAndID returns a baremetal private network API with zone and ID extracted from the state
func BaremetalPrivateNetworkAPIWithZoneAndID(m interface{}, id string) (*baremetal.PrivateNetworkAPI, zonal.ID, error) {
	baremetalPrivateNetworkAPI := baremetal.NewPrivateNetworkAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, zonal.ID{}, err
	}
	return baremetalPrivateNetworkAPI, zonal.NewID(zone, ID), nil
}

func expandBaremetalOptions(i interface{}) ([]*baremetal.ServerOption, error) {
	options := []*baremetal.ServerOption(nil)

	for _, op := range i.(*schema.Set).List() {
		rawOption := op.(map[string]interface{})
		option := &baremetal.ServerOption{}
		if optionExpiresAt, hasExpiresAt := rawOption["expires_at"]; hasExpiresAt {
			option.ExpiresAt = expandTimePtr(optionExpiresAt)
		}
		id := locality.ExpandID(rawOption["id"].(string))
		name := rawOption["name"].(string)

		option.ID = id
		option.Name = name

		options = append(options, option)
	}

	return options, nil
}

func expandBaremetalPrivateNetworks(pn interface{}) []string {
	privateNetworkIDs := make([]string, 0, len(pn.(*schema.Set).List()))

	for _, op := range pn.(*schema.Set).List() {
		rawPN := op.(map[string]interface{})
		id := locality.ExpandID(rawPN["id"].(string))
		privateNetworkIDs = append(privateNetworkIDs, id)
	}

	return privateNetworkIDs
}

func flattenBaremetalCPUs(cpus []*baremetal.CPU) interface{} {
	if cpus == nil {
		return nil
	}
	flattenedCPUs := []map[string]interface{}(nil)
	for _, cpu := range cpus {
		flattenedCPUs = append(flattenedCPUs, map[string]interface{}{
			"name":         cpu.Name,
			"core_count":   cpu.CoreCount,
			"frequency":    cpu.Frequency,
			"thread_count": cpu.ThreadCount,
		})
	}
	return flattenedCPUs
}

func flattenBaremetalDisks(disks []*baremetal.Disk) interface{} {
	if disks == nil {
		return nil
	}
	flattenedDisks := []map[string]interface{}(nil)
	for _, disk := range disks {
		flattenedDisks = append(flattenedDisks, map[string]interface{}{
			"type":     disk.Type,
			"capacity": disk.Capacity,
		})
	}
	return flattenedDisks
}

func flattenBaremetalMemory(memories []*baremetal.Memory) interface{} {
	if memories == nil {
		return nil
	}
	flattenedMemories := []map[string]interface{}(nil)
	for _, memory := range memories {
		flattenedMemories = append(flattenedMemories, map[string]interface{}{
			"type":      memory.Type,
			"capacity":  memory.Capacity,
			"frequency": memory.Frequency,
			"is_ecc":    memory.IsEcc,
		})
	}
	return flattenedMemories
}

func flattenBaremetalIPs(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}
	flattendIPs := []map[string]interface{}(nil)
	for _, ip := range ips {
		flattendIPs = append(flattendIPs, map[string]interface{}{
			"id":      ip.ID,
			"address": ip.Address.String(),
			"reverse": ip.Reverse,
			"version": ip.Version.String(),
		})
	}
	return flattendIPs
}

func flattenBaremetalIPv4s(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}
	flattendIPs := []map[string]interface{}(nil)
	for _, ip := range ips {
		if ip.Version == baremetal.IPVersionIPv4 {
			flattendIPs = append(flattendIPs, map[string]interface{}{
				"id":      ip.ID,
				"address": ip.Address.String(),
				"reverse": ip.Reverse,
				"version": ip.Version.String(),
			})
		}
	}
	return flattendIPs
}

func flattenBaremetalIPv6s(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}
	flattendIPs := []map[string]interface{}(nil)
	for _, ip := range ips {
		if ip.Version == baremetal.IPVersionIPv6 {
			flattendIPs = append(flattendIPs, map[string]interface{}{
				"id":      ip.ID,
				"address": ip.Address.String(),
				"reverse": ip.Reverse,
				"version": ip.Version.String(),
			})
		}
	}
	return flattendIPs
}

func flattenBaremetalOptions(zone scw.Zone, options []*baremetal.ServerOption) interface{} {
	if options == nil {
		return nil
	}
	flattenedOptions := []map[string]interface{}(nil)
	for _, option := range options {
		flattenedOptions = append(flattenedOptions, map[string]interface{}{
			"id":         zonal.NewID(zone, option.ID).String(),
			"expires_at": types.FlattenTime(option.ExpiresAt),
			"name":       option.Name,
		})
	}
	return flattenedOptions
}

func flattenBaremetalPrivateNetworks(region scw.Region, privateNetworks []*baremetal.ServerPrivateNetwork) interface{} {
	flattenedPrivateNetworks := []map[string]interface{}(nil)
	for _, privateNetwork := range privateNetworks {
		flattenedPrivateNetworks = append(flattenedPrivateNetworks, map[string]interface{}{
			"id":         regional.NewIDString(region, privateNetwork.PrivateNetworkID),
			"vlan":       types.FlattenUint32Ptr(privateNetwork.Vlan),
			"status":     privateNetwork.Status,
			"created_at": types.FlattenTime(privateNetwork.CreatedAt),
			"updated_at": types.FlattenTime(privateNetwork.UpdatedAt),
		})
	}
	return flattenedPrivateNetworks
}

func detachAllPrivateNetworkFromBaremetal(ctx context.Context, d *schema.ResourceData, m interface{}, serverID string) error {
	privateNetworkAPI, zone, err := baremetalPrivateNetworkAPIWithZone(d, m)
	if err != nil {
		return err
	}
	listPrivateNetwork, err := privateNetworkAPI.ListServerPrivateNetworks(&baremetal.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     zone,
		ServerID: &serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	for _, pn := range listPrivateNetwork.ServerPrivateNetworks {
		err := privateNetworkAPI.DeleteServerPrivateNetwork(&baremetal.PrivateNetworkAPIDeleteServerPrivateNetworkRequest{
			Zone:             zone,
			ServerID:         serverID,
			PrivateNetworkID: pn.PrivateNetworkID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}
	}

	_, err = waitForBaremetalServerPrivateNetwork(ctx, privateNetworkAPI, zone, serverID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return err
	}
	return nil
}

func waitForBaremetalServer(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := baremetalRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForBaremetalServerInstall(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := baremetalRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerInstall(&baremetal.WaitForServerInstallRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForBaremetalServerOptions(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := baremetalRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerOptions(&baremetal.WaitForServerOptionsRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForBaremetalServerPrivateNetwork(ctx context.Context, api *baremetal.PrivateNetworkAPI, zone scw.Zone, serverID string, timeout time.Duration) ([]*baremetal.ServerPrivateNetwork, error) {
	retryInterval := baremetalRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}
	serverPrivateNetwork, err := api.WaitForServerPrivateNetworks(&baremetal.WaitForServerPrivateNetworksRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return serverPrivateNetwork, err
}

func baremetalInstallServer(ctx context.Context, d *schema.ResourceData, baremetalAPI *baremetal.API, installServerRequest *baremetal.InstallServerRequest) error {
	installServerRequest.OsID = locality.ExpandID(d.Get("os"))
	installServerRequest.SSHKeyIDs = types.ExpandStrings(d.Get("ssh_key_ids"))

	_, err := baremetalAPI.InstallServer(installServerRequest, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func BaremetalFindOfferByID(ctx context.Context, baremetalAPI *baremetal.API, zone scw.Zone, offerID string) (*baremetal.Offer, error) {
	subscriptionPeriods := []baremetal.OfferSubscriptionPeriod{
		baremetal.OfferSubscriptionPeriodHourly,
		baremetal.OfferSubscriptionPeriodMonthly,
	}

	for _, subscriptionPeriod := range subscriptionPeriods {
		res, err := baremetalAPI.ListOffers(&baremetal.ListOffersRequest{
			Zone:               zone,
			SubscriptionPeriod: subscriptionPeriod,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for _, offer := range res.Offers {
			if offer.ID == offerID {
				return offer, nil
			}
		}
	}

	return nil, fmt.Errorf("offer %s not found in zone %s", offerID, zone)
}

func baremetalCompareOptions(slice1, slice2 []*baremetal.ServerOption) []*baremetal.ServerOption {
	var diff []*baremetal.ServerOption

	m := make(map[string]struct{}, len(slice1))
	for _, option := range slice1 {
		m[option.ID] = struct{}{}
	}
	// find the differences
	for _, option := range slice2 {
		if _, foundID := m[option.ID]; !foundID {
			diff = append(diff, option)
		} else if foundID {
			if _, foundExp := m[types.FlattenTime(option.ExpiresAt).(string)]; !foundExp {
				diff = append(diff, option)
			}
		}
	}
	return diff
}

// customDiffBaremetalPrivateNetworkOption checks that the private_network option has been set if there is a private_network
func customDiffBaremetalPrivateNetworkOption() func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
	return func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
		var isPrivateNetworkOption bool

		_, okPrivateNetwork := diff.GetOk("private_network")

		options, optionsExist := diff.GetOk("options")
		if optionsExist {
			opSpecs, err := expandBaremetalOptions(options)
			if err != nil {
				return err
			}

			for j := range opSpecs {
				// private network option ID
				if opSpecs[j].ID == "cd4158d7-2d65-49be-8803-c4b8ab6f760c" {
					isPrivateNetworkOption = true
				}
			}
		}

		if okPrivateNetwork && !isPrivateNetworkOption {
			return errors.New("private network option needs to be enabled in order to attach a private network")
		}

		return nil
	}
}

func baremetalPrivateNetworkSetHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if pnID, ok := m["id"]; ok {
		buf.WriteString(locality.ExpandID(pnID))
	}

	return StringHashcode(buf.String())
}
