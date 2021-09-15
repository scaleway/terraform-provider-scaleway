package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceIP().Schema)

	dsSchema["id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IP address",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"address"},
	}
	dsSchema["address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The IP address",
		ConflictsWith: []string{"id"},
		ValidateFunc:  validation.IsIPv4Address,
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstanceIPRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
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
			if is404Error(err) || is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		ID = res.IP.ID
	} else {
		_, ID, _ = parseLocalizedID(id.(string))
	}
	d.SetId(newZonedIDString(zone, ID))

	return resourceScalewayInstanceIPRead(ctx, d, meta)
}
