package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	baremetalServerWaitForTimeout   = 60 * time.Minute
	baremetalServerRetryFuncTimeout = baremetalServerWaitForTimeout + time.Minute // some RetryFunc are calling a WaitFor
	defaultBaremetalServerTimeout   = baremetalServerRetryFuncTimeout + time.Minute
	baremetalRetryInterval          = 5 * time.Second
)

// instanceAPIWithZone returns a new baremetal API and the zone for a Create request
func baremetalAPIWithZone(d *schema.ResourceData, m interface{}) (*baremetal.API, scw.Zone, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return baremetalAPI, zone, nil
}

// instanceAPIWithZoneAndID returns an baremetal API with zone and ID extracted from the state
func baremetalAPIWithZoneAndID(m interface{}, id string) (*baremetal.API, ZonedID, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, ZonedID{}, err
	}
	return baremetalAPI, newZonedID(zone, ID), nil
}

func expandBaremetalOptions(i interface{}) ([]*baremetal.ServerOption, error) {
	options := []*baremetal.ServerOption(nil)

	for _, op := range i.(*schema.Set).List() {
		rawOption := op.(map[string]interface{})
		option := &baremetal.ServerOption{}
		if optionExpiresAt, hasExpiresAt := rawOption["expires_at"]; hasExpiresAt {
			option.ExpiresAt = expandTimePtr(optionExpiresAt)
		}
		id := expandID(rawOption["id"].(string))
		option.ID = id
		options = append(options, option)
	}

	return options, nil
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

func flattenBaremetalOptions(zone scw.Zone, options []*baremetal.ServerOption) interface{} {
	if options == nil {
		return nil
	}
	flattenedOptions := []map[string]interface{}(nil)
	for _, option := range options {
		flattenedOptions = append(flattenedOptions, map[string]interface{}{
			"id":         newZonedID(zone, option.ID).String(),
			"expires_at": flattenTime(option.ExpiresAt),
		})
	}
	return flattenedOptions
}

func waitForBaremetalServer(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := baremetalRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
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
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
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
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerOptions(&baremetal.WaitForServerOptionsRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func baremetalInstallServer(ctx context.Context, d *schema.ResourceData, baremetalAPI *baremetal.API, installServerRequest *baremetal.InstallServerRequest) error {
	installServerRequest.OsID = expandID(d.Get("os"))
	installServerRequest.SSHKeyIDs = expandStrings(d.Get("ssh_key_ids"))

	_, err := baremetalAPI.InstallServer(installServerRequest, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func baremetalFindOfferByID(ctx context.Context, baremetalAPI *baremetal.API, zone scw.Zone, offerID string) (*baremetal.Offer, error) {
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

func baremetalCompareOptionIDsToAdd(modifiedOptions, currentOptions []*baremetal.ServerOption, zone scw.Zone) []*baremetal.ServerOption {
	var toAdd []*baremetal.ServerOption

	m := make(map[string]struct{}, len(currentOptions))
	for _, option := range currentOptions {
		m[option.ID] = struct{}{}
	}
	// find the differences
	for _, option := range modifiedOptions {
		if _, foundID := m[option.ID]; !foundID {
			toAdd = append(toAdd, option)
		} else if foundID {
			if _, foundExp := m[flattenTime(option.ExpiresAt).(string)]; !foundExp {
				toAdd = append(toAdd, option)
			}
		}
	}
	return toAdd
}

func baremetalCompareOptionIDsToDelete(modifiedOptions, currentOptions []*baremetal.ServerOption, zone scw.Zone) []*baremetal.ServerOption {
	var toDelete []*baremetal.ServerOption

	m := make(map[string]struct{}, len(modifiedOptions))
	for _, option := range modifiedOptions {
		m[option.ID] = struct{}{}
	}
	// find the differences
	for _, option := range currentOptions {
		if _, foundID := m[option.ID]; !foundID {
			toDelete = append(toDelete, option)
		} else if foundID {
			if _, foundExp := m[flattenTime(option.ExpiresAt).(string)]; !foundExp {
				toDelete = append(toDelete, option)
			}
		}
	}
	return toDelete
}
