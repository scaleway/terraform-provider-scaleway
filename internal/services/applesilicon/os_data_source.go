package applesilicon

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	ErrOSNotFound          = errors.New("no OS found")
	ErrOSNotFoundWithFilter = errors.New("no OS found with given filter")
)

func DataSourceOS() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceOSRead,
		SchemaFunc:  osSchema,
	}
}

func osSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Exact label of the desired image",
			ConflictsWith: []string{"os_id"},
		},
		"version": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
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
	}
}

func DataSourceOSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var osVersion, osName string

	osID, ok := d.GetOk("os_id")
	if ok {
		res, err := api.GetOS(&applesilicon.GetOSRequest{
			Zone: zone,
			OsID: osID.(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		osVersion = res.Version
		osName = res.Name
	} else {
		res, err := api.ListOS(&applesilicon.ListOSRequest{
			Zone: zone,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if res.TotalCount == 0 {
			return diag.FromErr(fmt.Errorf("%w: something went wrong when listing OS", ErrOSNotFound))
		}

		for _, os := range res.Os {
			if os.Name == d.Get("name") {
				osID, osVersion, osName = os.ID, os.Version, os.Name

				break
			}
		}

		if osID == "" {
			return diag.FromErr(fmt.Errorf(
				"%w with name=%q and version=%q in zone %s",
				ErrOSNotFoundWithFilter,
				d.Get("name"),
				d.Get("version"),
				zone,
			))
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
