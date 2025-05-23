package baremetal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	partitionSize = 20000000000
)

func DataEasyPartitioning() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		ReadContext:                       dataEasyPartitioningRead,
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
			"ext_4_mountpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/data",
				Description: "Mount point must be an absolute path with alphanumeric characters and underscores",
			},
			"json_partition": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The partitioning schema in json format",
			},
		},
	}
}

func dataEasyPartitioningRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	extraPart := d.Get("extra_partition").(bool)
	swap := d.Get("swap").(bool)

	if swap && !extraPart {
		jsonSchema, err := json.Marshal(defaultPartitioningSchema)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(fmt.Sprintf("%s-%s", offerID, osID))
		_ = d.Set("json_partition", string(jsonSchema))

		return nil
	}

	resizeRootPartition(defaultPartitioningSchema.Disks, swap, extraPart)
	defaultPartitioningSchema.Disks = handleSwapPartitions(defaultPartitioningSchema.Disks, extraPart, swap)

	mountpoint := d.Get("ext_4_mountpoint").(string)
	addExtraExt4Partition(mountpoint, defaultPartitioningSchema, extraPart)

	if !extraPart && !swap {
		defaultPartitioningSchema.Filesystems = defaultPartitioningSchema.Filesystems[:len(defaultPartitioningSchema.Filesystems)-1]
		defaultPartitioningSchema.Raids = defaultPartitioningSchema.Raids[:len(defaultPartitioningSchema.Raids)-1]
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

func handleSwapPartitions(originalDisks []*baremetal.SchemaDisk, withExtraPartition bool, swap bool) []*baremetal.SchemaDisk {
	if swap {
		return originalDisks
	}

	result := make([]*baremetal.SchemaDisk, 0)

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
					p.Size = partitionSize
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

func addExtraExt4Partition(mountpoint string, defaultPartitionSchema *baremetal.Schema, extraPart bool) {
	if !extraPart {
		return
	}

	for _, disk := range defaultPartitionSchema.Disks {
		partIndex := uint32(len(disk.Partitions)) + 1
		data := &baremetal.SchemaPartition{
			Label:                baremetal.SchemaPartitionLabel("data"),
			Number:               partIndex,
			Size:                 0,
			UseAllAvailableSpace: true,
		}
		disk.Partitions = append(disk.Partitions, data)
	}

	filesystem := &baremetal.SchemaFilesystem{
		Device:     "/dev/md2",
		Format:     "ext4",
		Mountpoint: mountpoint,
	}
	defaultPartitionSchema.Filesystems = append(defaultPartitionSchema.Filesystems, filesystem)
}

func resizeRootPartition(originalDisks []*baremetal.SchemaDisk, withSwap bool, withExtraPartition bool) {
	for _, disk := range originalDisks {
		for _, partition := range disk.Partitions {
			if partition.Label == "root" {
				if !withSwap && !withExtraPartition {
					partition.Size = 0
					partition.UseAllAvailableSpace = true
				}

				if withExtraPartition {
					partition.Size = partitionSize
				}
			}
		}
	}
}
