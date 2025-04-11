package baremetal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataEasyPartitioning() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataEasyPartitioningRead,
		Schema: map[string]*schema.Schema{
			"offer_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server offer",
			},
			"os_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The base image of the server",
				DiffSuppressFunc: dsf.Locality,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"swap": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "set swap partition",
			},
			"extra_partition": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "set extra ext_4 partition",
			},
			"ext_4_mountpoint": { //TODO change to mount point
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/hello",
				Description: "Mount point must be an absolute path with alphanumeric characters and underscores",
			},
			"json_partition": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The partitioning schema in json format",
			},
			"disks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"partitions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"label": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"number": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"size": {
										Type:     schema.TypeString, // scw.Size implements String()
										Computed: true,
									},
									"use_all_available_space": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"raids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"devices": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"filesystems": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mountpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func removeSwap(originalDisks []*baremetal.SchemaDisk, withExtraPartition bool) []*baremetal.SchemaDisk {
	var result []*baremetal.SchemaDisk

	for _, disk := range originalDisks {
		i := 1
		newPartitions := []*baremetal.SchemaPartition{}
		for _, p := range disk.Partitions {
			if p.Label == "swap" {
				continue
			}
			if p.Label == "root" {
				if !withExtraPartition {
					p.Size = 0
					p.UseAllAvailableSpace = true
				} else {
					p.Size = 20000000000
				}
			}
			p.Number = uint32(i)
			i++
			newPartitions = append(newPartitions, p)
		}
		result = append(result, &baremetal.SchemaDisk{
			Device:     disk.Device,
			Partitions: newPartitions,
		})
	}
	return result
}

func addExtraPartition(mountpoint string, newDisksSchema []*baremetal.SchemaDisk, defaultPartitionSchema *baremetal.Schema) *baremetal.Schema {
	raidDevices := []string{}

	for _, disk := range newDisksSchema {
		partIndex := uint32(len(disk.Partitions)) + 1
		deviceIndex := partIndex + 1
		data := &baremetal.SchemaPartition{
			Label:                baremetal.SchemaPartitionLabel("data"),
			Number:               partIndex,
			Size:                 0,
			UseAllAvailableSpace: true,
		}
		disk.Partitions = append(disk.Partitions, data)

		device := fmt.Sprintf("%sp%d", disk.Device, deviceIndex)
		raidDevices = append(raidDevices, device)
		deviceIndex--
	}

	filesystem := &baremetal.SchemaFilesystem{
		Device:     "/dev/md2",
		Format:     "ext4",
		Mountpoint: mountpoint,
	}
	defaultPartitionSchema.Filesystems = append(defaultPartitionSchema.Filesystems, filesystem)

	raid := &baremetal.SchemaRAID{
		Name:    "/dev/md2",
		Level:   baremetal.SchemaRAIDLevelRaidLevel1,
		Devices: raidDevices,
	}
	defaultPartitionSchema.Raids = append(defaultPartitionSchema.Raids, raid)
	defaultPartitionSchema.Disks = newDisksSchema

	return defaultPartitionSchema
}

func manageRootSize(originalDisks []*baremetal.SchemaDisk, withSwap bool, withExtraPartition bool) {
	for _, disk := range originalDisks {
		for _, partition := range disk.Partitions {
			if partition.Label == "root" {
				if !withSwap && !withExtraPartition {
					partition.Size = 0
					partition.UseAllAvailableSpace = true
				}
				if withExtraPartition {
					partition.Size = 20000000000
				}
			}
		}
	}
}

func flattenDisksSchema(disks []*baremetal.SchemaDisk) []map[string]interface{} {
	var out []map[string]interface{}
	for _, d := range disks {
		if d == nil {
			continue
		}

		parts := make([]map[string]interface{}, 0, len(d.Partitions))
		for _, p := range d.Partitions {
			if p == nil {
				continue
			}
			parts = append(parts, map[string]interface{}{
				"label":                   string(p.Label),
				"number":                  int(p.Number),
				"size":                    p.Size.String(),
				"use_all_available_space": p.UseAllAvailableSpace,
			})
		}

		out = append(out, map[string]interface{}{
			"device":     d.Device,
			"partitions": parts,
		})
	}
	return out
}

func flattenRaids(raids []*baremetal.SchemaRAID) []map[string]interface{} {
	var out []map[string]interface{}
	for _, r := range raids {
		if r == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"name":    r.Name,
			"level":   string(r.Level),
			"devices": r.Devices,
		})
	}
	return out
}

func flattenFilesystems(fsList []*baremetal.SchemaFilesystem) []map[string]interface{} {
	var out []map[string]interface{}
	for _, fs := range fsList {
		if fs == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"device":     fs.Device,
			"format":     string(fs.Format),
			"mountpoint": fs.Mountpoint,
		})
	}
	return out
}

func dataEasyPartitioningRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, fallBackZone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	osID := d.Get("os_id").(string)

	os, err := api.GetOS(&baremetal.GetOSRequest{
		Zone: fallBackZone,
		OsID: osID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if !os.CustomPartitioningSupported {
		return diag.FromErr(fmt.Errorf("custom partitioning is not supported with this OS"))
	}

	offerID := d.Get("offer_id").(string)

	offer, err := api.GetOffer(&baremetal.GetOfferRequest{
		Zone:    fallBackZone,
		OfferID: offerID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if !isOSCompatible(offer, os) {
		return diag.FromErr(fmt.Errorf("os and offer are not compatible"))
	}

	defaultPartitioningSchema, err := api.GetDefaultPartitioningSchema(&baremetal.GetDefaultPartitioningSchemaRequest{
		Zone:    fallBackZone,
		OfferID: offerID,
		OsID:    osID,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	extraPart := d.Get("extra_partition").(bool)
	swap := d.Get("swap").(bool)

	if swap && !extraPart {
		jsonSchema, _ := json.Marshal(defaultPartitioningSchema)
		d.SetId(fmt.Sprintf("%s-%s", offerID, osID))
		_ = d.Set("json_partition", string(jsonSchema))
		_ = d.Set("disks", flattenDisksSchema(defaultPartitioningSchema.Disks))
		_ = d.Set("raids", flattenRaids(defaultPartitioningSchema.Raids))
		_ = d.Set("filesystems", flattenFilesystems(defaultPartitioningSchema.Filesystems))

		return nil
	}

	manageRootSize(defaultPartitioningSchema.Disks, swap, extraPart)

	var newDiskSchema []*baremetal.SchemaDisk
	if !swap {
		newDiskSchema = removeSwap(defaultPartitioningSchema.Disks, extraPart)
	}

	if newDiskSchema == nil {
		newDiskSchema = defaultPartitioningSchema.Disks
	}

	var newCustomPartition *baremetal.Schema
	if extraPart {
		mountpoint := d.Get("ext_4_mountpoint").(string)
		newCustomPartition = addExtraPartition(mountpoint, newDiskSchema, defaultPartitioningSchema)
	}

	err = api.ValidatePartitioningSchema(&baremetal.ValidatePartitioningSchemaRequest{
		Zone:               fallBackZone,
		OfferID:            offerID,
		OsID:               osID,
		PartitioningSchema: defaultPartitioningSchema,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	jsonSchema, err := json.Marshal(newCustomPartition)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s-%s", offerID, osID))
	_ = d.Set("json_partition", string(jsonSchema))
	_ = d.Set("disks", flattenDisksSchema(newCustomPartition.Disks))
	_ = d.Set("raids", flattenRaids(newCustomPartition.Raids))
	_ = d.Set("filesystems", flattenFilesystems(newCustomPartition.Filesystems))

	return nil
}
