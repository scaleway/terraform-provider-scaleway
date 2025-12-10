package baremetal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	partitionSize     = 20000000000
	defaultMountpoint = "/data"
	md0               = "/dev/md0"
	md1               = "/dev/md1"
	md2               = "/dev/md2"
	ext4              = "ext4"
	raidLevel1        = "raid_level_1"
	nvme0p2           = "/dev/nvme0n1p2"
	nvme0p3           = "/dev/nvme0n1p3"
	nvme1p1           = "/dev/nvme1n1p1"
	nvme1p2           = "/dev/nvme1n1p2"
	uefi              = "uefi"
	swap              = "swap"
	root              = "root"
)

func DataPartitionSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataPartitionSchemaRead,
		SchemaFunc:  partitionSchema,
	}
}

func partitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"ext_4_mountpoint": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      defaultMountpoint,
			ValidateFunc: validation.StringInSlice([]string{"/data", "/home"}, false),
			Description:  "Mount point must be an absolute path",
		},
		"json_partition": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The partitioning schema in json format",
		},
	}
}

func dataPartitionSchemaRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, fallBackZone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	osID := zonal.ExpandID(d.Get("os_id").(string))

	os, err := api.GetOS(&baremetal.GetOSRequest{
		Zone: fallBackZone,
		OsID: osID.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if !os.CustomPartitioningSupported {
		return diag.FromErr(errors.New("custom partitioning is not supported with this OS"))
	}

	offerID := zonal.ExpandID(d.Get("offer_id").(string))

	offer, err := api.GetOffer(&baremetal.GetOfferRequest{
		Zone:    fallBackZone,
		OfferID: offerID.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if !isOSCompatible(offer, os) {
		return diag.FromErr(errors.New("OS and offer are not compatible"))
	}

	defaultPartitioningSchema, err := api.GetDefaultPartitioningSchema(&baremetal.GetDefaultPartitioningSchemaRequest{
		Zone:    fallBackZone,
		OfferID: offerID.ID,
		OsID:    osID.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	hasSwap := d.Get("swap").(bool)
	if !hasSwap {
		removeSwap(defaultPartitioningSchema.Disks)
		updateRaidRemoveSwap(defaultPartitioningSchema)
	}

	mountpoint := d.Get("ext_4_mountpoint").(string)
	_, hasExtraPartition := d.GetOk("extra_partition")

	if hasExtraPartition {
		addExtraExt4Partition(mountpoint, defaultPartitioningSchema)
		updateRaidNewPartition(defaultPartitioningSchema)
	}

	if !hasSwap || hasExtraPartition {
		updateSizeRoot(defaultPartitioningSchema.Disks, hasExtraPartition)
	}

	err = api.ValidatePartitioningSchema(&baremetal.ValidatePartitioningSchemaRequest{
		Zone:               fallBackZone,
		OfferID:            offerID.ID,
		OsID:               osID.ID,
		PartitioningSchema: defaultPartitioningSchema,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	jsonSchema, err := json.Marshal(defaultPartitioningSchema)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s-%s", offerID, osID))

	jsonSchemaStr := string(jsonSchema)

	_ = d.Set("json_partition", jsonSchemaStr)

	return nil
}

func updateRaidRemoveSwap(partitionSchema *baremetal.Schema) {
	raidSchema := []*baremetal.SchemaRAID{
		{
			Name:  md0,
			Level: raidLevel1,
			Devices: []string{
				nvme0p2,
				nvme1p1,
			},
		},
		{
			Name:  md1,
			Level: raidLevel1,
			Devices: []string{
				nvme0p3,
				nvme1p2,
			},
		},
	}
	partitionSchema.Raids = raidSchema
}

func updateRaidNewPartition(partitionSchema *baremetal.Schema) {
	lenDisk0Partition := len(partitionSchema.Disks[0].Partitions)
	lenDisk1Partition := len(partitionSchema.Disks[1].Partitions)
	raid := &baremetal.SchemaRAID{
		Name:  md2,
		Level: raidLevel1,
		Devices: []string{
			fmt.Sprintf("%sp%d", partitionSchema.Disks[0].Device, lenDisk0Partition),
			fmt.Sprintf("%sp%d", partitionSchema.Disks[1].Device, lenDisk1Partition),
		},
	}
	partitionSchema.Raids = append(partitionSchema.Raids, raid)
}

func addExtraExt4Partition(mountpoint string, defaultPartitionSchema *baremetal.Schema) {
	label := strings.TrimPrefix(mountpoint, "/")

	for _, disk := range defaultPartitionSchema.Disks {
		partIndex := uint32(len(disk.Partitions)) + 1
		data := &baremetal.SchemaPartition{
			Label:                baremetal.SchemaPartitionLabel(label),
			Number:               partIndex,
			Size:                 0,
			UseAllAvailableSpace: true,
		}
		disk.Partitions = append(disk.Partitions, data)
	}

	filesystem := &baremetal.SchemaFilesystem{
		Device:     md2,
		Format:     ext4,
		Mountpoint: mountpoint,
	}
	defaultPartitionSchema.Filesystems = append(defaultPartitionSchema.Filesystems, filesystem)
}

func updateSizeRoot(originalDisks []*baremetal.SchemaDisk, hasExtraPartition bool) {
	for _, disk := range originalDisks {
		for _, partition := range disk.Partitions {
			if partition.Label == root {
				partition.Size = 0
				partition.UseAllAvailableSpace = true

				if hasExtraPartition {
					partition.Size = partitionSize
					partition.UseAllAvailableSpace = false
				}
			}
		}
	}
}

func removeSwap(originalDisks []*baremetal.SchemaDisk) {
	for _, disk := range originalDisks {
		newPartitions := make([]*baremetal.SchemaPartition, 0, len(disk.Partitions))

		for _, partition := range disk.Partitions {
			if partition.Label == swap {
				continue
			}

			if partition.Label != uefi {
				partition.Number--
			}

			newPartitions = append(newPartitions, partition)
		}

		disk.Partitions = newPartitions
	}
}
