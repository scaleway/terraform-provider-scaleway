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
	"strings"
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
	defaultPartitionSchema.Disks = newDisksSchema

	filesystem := &baremetal.SchemaFilesystem{
		Device:     "/dev/md2",
		Format:     "ext4",
		Mountpoint: mountpoint,
	}
	defaultPartitionSchema.Filesystems = append(defaultPartitionSchema.Filesystems, filesystem)

	//raid := &baremetal.SchemaRAID{
	//	Name:    "/dev/md2",
	//	Level:   baremetal.SchemaRAIDLevelRaidLevel1,
	//	Devices: raidDevices,
	//}
	//defaultPartitionSchema.Raids = append(defaultPartitionSchema.Raids, raid)

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
		//_ = d.Set("disks", flattenDisksSchema(defaultPartitioningSchema.Disks))
		//_ = d.Set("raids", flattenRaids(defaultPartitioningSchema.Raids))
		//_ = d.Set("filesystems", flattenFilesystems(defaultPartitioningSchema.Filesystems))

		return nil
	}

	manageRootSize(defaultPartitioningSchema.Disks, swap, extraPart)

	var newDiskSchema []*baremetal.SchemaDisk
	if !swap {
		newDiskSchema = removeSwap(defaultPartitioningSchema.Disks, extraPart)
	} else {
		newDiskSchema = defaultPartitioningSchema.Disks
	}

	var newCustomPartition *baremetal.Schema
	if extraPart {
		mountpoint := d.Get("ext_4_mountpoint").(string)
		newCustomPartition = addExtraPartition(mountpoint, newDiskSchema, defaultPartitioningSchema)
	} else {
		newCustomPartition = defaultPartitioningSchema
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
	jsonSchemaStr := string(jsonSchema)
	strings.ReplaceAll(jsonSchemaStr, "\"", "\\\"")
	_ = d.Set("json_partition", jsonSchemaStr)
	//_ = d.Set("disks", flattenDisksSchema(newCustomPartition.Disks))
	//_ = d.Set("raids", flattenRaids(newCustomPartition.Raids))
	//_ = d.Set("filesystems", flattenFilesystems(newCustomPartition.Filesystems))

	return nil
}
