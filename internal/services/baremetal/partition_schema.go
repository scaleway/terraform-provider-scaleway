package baremetal

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func PartitioningSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Configuration for partitioning schema.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disks": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "List of disks with partitions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"device": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Device name (e.g., /dev/nvme0n1).",
							},
							"partitions": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Partitions for the disk.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"label": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Partition label.",
										},
										"number": {
											Type:        schema.TypeInt,
											Required:    true,
											Description: "Partition number.",
										},
										"size": {
											Type:        schema.TypeInt,
											Required:    true,
											Description: "Partition size in bytes.",
										},
									},
								},
							},
						},
					},
				},
				"raids": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "RAID configurations.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the RAID device.",
							},
							"level": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "RAID level.",
							},
							"devices": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
								Description: "Devices in the RAID.",
							},
						},
					},
				},
				"filesystems": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Filesystem configurations.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"device": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Device name for the filesystem.",
							},
							"format": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Filesystem format.",
							},
							"mountpoint": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Mountpoint for the filesystem.",
							},
						},
					},
				},
				"zfs": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "ZFS configurations.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"pools": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "List of ZFS pools.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Name of the ZFS pool.",
										},
										"type": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Type of the ZFS pool (e.g., mirror, raidz).",
										},
										"devices": {
											Type:     schema.TypeList,
											Required: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
											Description: "Devices in the ZFS pool.",
										},
										"options": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
											Description: "Options for the ZFS pool.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
