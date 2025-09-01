package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceIP().Schema)

	dsSchema["id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IP address",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"address"},
	}
	dsSchema["address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The IP address",
		ConflictsWith: []string{"id"},
		ValidateFunc:  validation.IsIPv4Address,
	}

	return &schema.Resource{
		ReadContext: DataSourceInstanceIPRead,

		Schema: dsSchema,
	}
}

func DataSourceInstanceIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := d.GetOk("id")

	var ID string

	if !ok {
		res, err := instanceAPI.GetIP(&instance.GetIPRequest{
			IP:   d.Get("address").(string),
			Zone: zone,
		}, scw.WithContext(ctx))
		if err != nil {
			// We check for 403 because instance API returns 403 for a deleted IP
			if httperrors.Is404(err) || httperrors.Is403(err) {
				d.SetId("")

				return nil
			}

			return diag.FromErr(err)
		}

		ID = res.IP.ID
	} else {
		_, ID, _ = locality.ParseLocalizedID(id.(string))
	}

	d.SetId(zonal.NewIDString(zone, ID))

	return ResourceInstanceIPRead(ctx, d, m)
}
