package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceIP().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "address")
	dsSchema["id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the IP address",
		ValidateFunc: validationUUIDWithLocality(),
	}
	dsSchema["address"].ConflictsWith = []string{"id"}

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

	addr, ok := d.GetOk("address")

	var search string
	if ok {
		search = addr.(string)
	} else {
		_, search, _ = parseLocalizedID(d.Get("id").(string))
	}
	res, err := instanceAPI.GetIP(&instance.GetIPRequest{
		IP:   search,
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
	d.SetId(newZonedIDString(zone, res.IP.ID))

	return resourceScalewayInstanceIPRead(ctx, d, meta)
}
