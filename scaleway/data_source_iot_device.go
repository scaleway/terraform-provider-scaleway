package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func dataSourceScalewayIotDevice() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIotDevice().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"device_id"}
	dsSchema["hub_id"].Optional = true
	dsSchema["device_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IOT Device",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIotDeviceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIotDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := iotAPIWithRegion(d, meta)
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
			Name:   expandStringPtr(deviceName),
			HubID:  expandStringPtr(hubID),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundDevice, err := findExact(
			res.Devices,
			func(s *iot.Device) bool { return s.Name == deviceName },
			deviceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		deviceID = foundDevice.ID
	}

	regionalID := datasourceNewRegionalID(deviceID, region)
	d.SetId(regionalID)
	err = d.Set("device_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayIotDeviceRead(ctx, d, meta)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return diag.Errorf("IOT Device not found (%s)", regionalID)
	}
	return nil
}
