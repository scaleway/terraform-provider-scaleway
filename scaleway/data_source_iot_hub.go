package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIotHub() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIotHub().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"hub_id"}
	dsSchema["hub_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IOT Hub",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIotHubRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIotHubRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := iotAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	hubID, ok := d.GetOk("hub_id")
	if !ok {
		res, err := api.ListHubs(&iot.ListHubsRequest{
			Region:    region,
			ProjectID: expandStringPtr(d.Get("project_id")),
			Name:      expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, hub := range res.Hubs {
			if hub.Name == d.Get("name").(string) {
				if hubID != "" {
					return diag.Errorf("more than 1 hub with the same name %s", d.Get("name"))
				}
				hubID = hub.ID
			}
		}
		if hubID == "" {
			return diag.Errorf("no hub found with the name %s", d.Get("name"))
		}
	}

	regionalID := datasourceNewRegionalizedID(hubID, region)
	d.SetId(regionalID)
	err = d.Set("hub_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayIotHubRead(ctx, d, meta)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return diag.Errorf("IOT Hub not found (%s)", regionalID)
	}
	return nil
}
