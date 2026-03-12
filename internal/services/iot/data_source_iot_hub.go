package iot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceHub() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceHub().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"hub_id"}
	dsSchema["hub_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IOT Hub",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceIotHubRead,
		Schema:      dsSchema,
	}
}

func DataSourceIotHubRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

		foundHub, err := datasource.FindExact(
			res.Hubs,
			func(s *iot.Hub) bool { return s.Name == hubName },
			hubName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		hubID = foundHub.ID
	}

	regionalID := datasource.NewRegionalID(hubID, region)
	d.SetId(regionalID)

	err = d.Set("hub_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceIotHubRead(ctx, d, m)
	if diags != nil {
		return diags
	}

	if d.Id() == "" {
		return diag.Errorf("IOT Hub not found (%s)", regionalID)
	}

	return nil
}
