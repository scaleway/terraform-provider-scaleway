package baremetal

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	baremetalV3 "github.com/scaleway/scaleway-sdk-go/api/baremetal/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandOptions(i interface{}) ([]*baremetal.ServerOption, error) {
	options := []*baremetal.ServerOption(nil)

	for _, op := range i.(*schema.Set).List() {
		rawOption := op.(map[string]interface{})
		option := &baremetal.ServerOption{}

		if optionExpiresAt, hasExpiresAt := rawOption["expires_at"]; hasExpiresAt {
			option.ExpiresAt = types.ExpandTimePtr(optionExpiresAt)
		}

		id := locality.ExpandID(rawOption["id"].(string))
		name := rawOption["name"].(string)

		option.ID = id
		option.Name = name

		options = append(options, option)
	}

	return options, nil
}

func expandPrivateNetworks(pn interface{}) map[string]*[]string {
	privateNetworks := make(map[string]*[]string)

	for _, op := range pn.(*schema.Set).List() {
		rawPN := op.(map[string]interface{})
		id := locality.ExpandID(rawPN["id"].(string))

		ipamIPIDs := &[]string{}

		if ipamIPs, ok := rawPN["ipam_ip_ids"]; ok && ipamIPs != nil {
			ipamIPsList := ipamIPs.([]interface{})
			if len(ipamIPsList) > 0 {
				ips := make([]string, len(ipamIPsList))
				for i, ip := range ipamIPsList {
					ips[i] = locality.ExpandID(ip.(string))
				}

				ipamIPIDs = &ips
			}
		}

		privateNetworks[id] = ipamIPIDs
	}

	return privateNetworks
}

func flattenCPUs(cpus []*baremetal.CPU) interface{} {
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

func flattenDisks(disks []*baremetal.Disk) interface{} {
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

func flattenMemory(memories []*baremetal.Memory) interface{} {
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

func flattenIPs(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}

	flatIPs := []map[string]interface{}(nil)
	for _, ip := range ips {
		flatIPs = append(flatIPs, map[string]interface{}{
			"id":      ip.ID,
			"address": ip.Address.String(),
			"reverse": ip.Reverse,
			"version": ip.Version.String(),
		})
	}

	return flatIPs
}

func flattenIPv4s(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}

	flatIPs := []map[string]interface{}(nil)

	for _, ip := range ips {
		if ip.Version == baremetal.IPVersionIPv4 {
			flatIPs = append(flatIPs, map[string]interface{}{
				"id":      ip.ID,
				"address": ip.Address.String(),
				"reverse": ip.Reverse,
				"version": ip.Version.String(),
			})
		}
	}

	return flatIPs
}

func flattenIPv6s(ips []*baremetal.IP) interface{} {
	if ips == nil {
		return nil
	}

	flatIPs := []map[string]interface{}(nil)

	for _, ip := range ips {
		if ip.Version == baremetal.IPVersionIPv6 {
			flatIPs = append(flatIPs, map[string]interface{}{
				"id":      ip.ID,
				"address": ip.Address.String(),
				"reverse": ip.Reverse,
				"version": ip.Version.String(),
			})
		}
	}

	return flatIPs
}

func flattenOptions(zone scw.Zone, options []*baremetal.ServerOption) interface{} {
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

func flattenPrivateNetworks(region scw.Region, privateNetworks []*baremetalV3.ServerPrivateNetwork) interface{} {
	flattenedPrivateNetworks := []map[string]interface{}(nil)
	for _, privateNetwork := range privateNetworks {
		flattenedPrivateNetworks = append(flattenedPrivateNetworks, map[string]interface{}{
			"id":          regional.NewIDString(region, privateNetwork.PrivateNetworkID),
			"mapping_id":  regional.NewIDString(region, privateNetwork.ID),
			"ipam_ip_ids": regional.NewIDStrings(region, privateNetwork.IpamIPIDs),
			"vlan":        types.FlattenUint32Ptr(privateNetwork.Vlan),
			"status":      privateNetwork.Status,
			"created_at":  types.FlattenTime(privateNetwork.CreatedAt),
			"updated_at":  types.FlattenTime(privateNetwork.UpdatedAt),
		})
	}

	return flattenedPrivateNetworks
}
