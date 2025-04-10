package baremetal

import (
	"context"
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
			"ext_4": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "set extra ext_4 partition",
			},
			"ext_4_name": { //TODO change to mount point
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/data",
				Description: "Mount point must be an absolute path with alphanumeric characters and underscores",
			},
			"custom_partition": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The partitioning schema in json format",
			},
		},
	}
}

func removeSwap(defaultDisks []*baremetal.SchemaDisk, extraPartition bool) []*baremetal.SchemaDisk {
	var swapSize scw.Size
	var newPartition []*baremetal.SchemaPartition
	var newDisks []*baremetal.SchemaDisk
	var disk *baremetal.SchemaDisk

	for _, oldDisk := range defaultDisks {
		for _, partition := range oldDisk.Partitions {
			if partition.Label == "swap" {
				swapSize = partition.Size
				continue
			}
			if partition.Label == "boot" && !extraPartition {
				partition.Size += swapSize
			} else if partition.Label == "boot" && extraPartition {
				partition.Size = 20000000000
			}
			newPartition = append(newPartition, partition)
		}
		disk.Device = oldDisk.Device
		disk.Partitions = newPartition
		newDisks = append(newDisks, oldDisk)
	}
	return newDisks
}

"raids": [
{
"name": "/dev/md2",
"level": "raid_level_1",
"devices": [
"/dev/nvme0n1p5",
"/dev/nvme1n1p4"
]
}
],
"filesystems": [
{
"device": "/dev/md2",
"format": "ext4",
"mountpoint": "/home"
}
],

{
"label": "data",
"number": 4,
"size": 0,
"use_all_available_space": true
}

func addExtraPartition(name string, extraPartition []*baremetal.SchemaDisk, defaultPartitionSchema *baremetal.Schema) *baremetal.Schema {
	_, label, _ := strings.Cut(name, "/")
	data := &baremetal.SchemaPartition{
		Label:                "",
		Number:               0,
		Size:                 0,
		UseAllAvailableSpace: false,
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

	var newDiskSchema []*baremetal.SchemaDisk
	if swap := d.Get("swap"); !swap.(bool) {
		newDiskSchema = removeSwap(defaultPartitioningSchema.Disks, extraPart)
	}

	var newCustomPartition []*baremetal.Schema
	if extraPart {
		name := d.Get("ext_4_name").(string)
		newCustomPartition = addExtraPartition(name, newDiskSchema, defaultPartitioningSchema)
	}

	defaultPartitioningSchema.Disks = append(defaultPartitioningSchema.Disks)
	//TODO checker si offer custom partitoning2l;
	//TODO checker si offer et os compatible
	//TODO get default partitioning
	//TODO remove swap and increase boot size
	//TODO
	//TODO unmarshall
	//TODO replacer les valeurs
	//TODO marshal

	return nil
}
