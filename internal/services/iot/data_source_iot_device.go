package iot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDevice() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDevice().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"device_id"}
	dsSchema["hub_id"].Optional = true
	dsSchema["device_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IOT Device",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceIotDeviceRead,
		Schema:      dsSchema,
	}
}

func DataSourceIotDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := iotAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	deviceID, ok := d.GetOk("device_id")
	if !ok {
		hubID, hubIDExists := d.GetOk("hub_id")
		if hubIDExists {
			_, hubID, err = regional.ParseID(hubID.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		deviceName := d.Get("name").(string)

		res, err := api.ListDevices(&iot.ListDevicesRequest{
			Region: region,
			Name:   types.ExpandStringPtr(deviceName),
			HubID:  types.ExpandStringPtr(hubID),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundDevice, err := datasource.FindExact(
			res.Devices,
			func(s *iot.Device) bool { return s.Name == deviceName },
			deviceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		deviceID = foundDevice.ID
	}

	regionalID := datasource.NewRegionalID(deviceID, region)
	d.SetId(regionalID)

	err = d.Set("device_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceIotDeviceRead(ctx, d, m)
	if diags != nil {
		return diags
	}

	if d.Id() == "" {
		return diag.Errorf("IOT Device not found (%s)", regionalID)
	}

	return nil
}
