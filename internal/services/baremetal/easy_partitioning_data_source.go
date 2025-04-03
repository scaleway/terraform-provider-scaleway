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
				Description: "set ext_4 partition",
			},
			"ext_4_name": {
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

	print(defaultPartitioningSchema)
	//TODO checker si offer custom partitoning
	//TODO checker si offer et os compatible
	//TODO get default partitioning
	//TODO unmarshall
	//TODO replacer les valeurs
	//TODO marshal

	return nil
}
