package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
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
			_, hubID, err = parseRegionalID(hubID.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
		res, err := api.ListDevices(&iot.ListDevicesRequest{
			Region: region,
			Name:   expandStringPtr(d.Get("name")),
			HubID:  expandStringPtr(hubID),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		for _, device := range res.Devices {
			if device.Name == d.Get("name").(string) {
				if deviceID != "" {
					return diag.Errorf("more than 1 device with the same name %s", d.Get("name"))
				}
				deviceID = device.ID
			}
		}
		if deviceID == "" {
			return diag.Errorf("no device found with the name %s", d.Get("name"))
		}
	}

	regionalID := datasourceNewRegionalizedID(deviceID, region)
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
