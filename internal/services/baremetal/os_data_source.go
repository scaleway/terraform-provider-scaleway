package baremetal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceOS() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		ReadContext:                       DataSourceOSRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact label of the desired image",
				ConflictsWith: []string{"os_id"},
			},
			"version": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Version string of the desired OS",
				ConflictsWith: []string{"os_id"},
			},
			"os_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the os",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"name"},
			},
			"zone": zonal.Schema(),
		},
	}
}

func DataSourceOSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var osVersion, osName string

	osID, ok := d.GetOk("os_id")
	if ok {
		// We fetch the name and version using the os id
		osID = d.Get("os_id")

		res, err := api.GetOS(&baremetal.GetOSRequest{
			Zone: zone,
			OsID: osID.(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		osVersion = res.Version
		osName = res.Name
	} else {
		// Get server by zone and name.
		res, err := api.ListOS(&baremetal.ListOSRequest{
			Zone: zone,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if len(res.Os) == 0 {
			return diag.FromErr(fmt.Errorf("no os found with the name %s", d.Get("name")))
		}

		for _, os := range res.Os {
			if os.Name == d.Get("name") && os.Version == d.Get("version") {
				osID, osVersion, osName = os.ID, os.Version, os.Name

				break
			}
		}
	}

	zoneID := datasource.NewZonedID(osID, zone)
	d.SetId(zoneID)

	_ = d.Set("os_id", zoneID)
	_ = d.Set("zone", zone)
	_ = d.Set("name", osName)
	_ = d.Set("version", osVersion)

	return nil
}
