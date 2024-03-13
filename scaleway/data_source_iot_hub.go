package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayIotHub() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIotHub().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"hub_id"}
	dsSchema["hub_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IOT Hub",
		ConflictsWith: []string{"name"},
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIotHubRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIotHubRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := iotAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	hubID, ok := d.GetOk("hub_id")
	if !ok {
		hubName := d.Get("name").(string)
		res, err := api.ListHubs(&iot.ListHubsRequest{
			Region:    region,
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
			Name:      types.ExpandStringPtr(hubName),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundHub, err := findExact(
			res.Hubs,
			func(s *iot.Hub) bool { return s.Name == hubName },
			hubName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		hubID = foundHub.ID
	}

	regionalID := datasourceNewRegionalID(hubID, region)
	d.SetId(regionalID)
	err = d.Set("hub_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayIotHubRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return diag.Errorf("IOT Hub not found (%s)", regionalID)
	}
	return nil
}
