package iot

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceHub() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceIotHubCreate,
		ReadContext:   ResourceIotHubRead,
		UpdateContext: ResourceIotHubUpdate,
		DeleteContext: ResourceIotHubDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultIoTHubTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    hubSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func hubSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable the hub or not",
			Default:     true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the hub",
		},
		"product_plan": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "The product plan of the hub",
			ValidateDiagFunc: verify.ValidateEnum[iot.HubProductPlan](),
		},
		"disable_events": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to enable the hub events or not",
		},
		"events_topic_prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Topic prefix for the hub events",
			Default:     "$SCW/events",
		},
		"hub_ca": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom user provided certificate authority",
		},
		"hub_ca_challenge": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Challenge certificate for the user provided hub CA",
			RequiredWith: []string{"hub_ca"},
		},
		"device_auto_provisioning": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Wether to enable the device auto provisioning or not",
		},

		// Computed elements
		"region":          regional.Schema(),
		"organization_id": account.OrganizationIDSchema(),
		"project_id":      account.ProjectIDSchema(),
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the IoT Hub",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the IoT Hub",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the hub",
		},
		"endpoint": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The endpoint to connect the devices to",
		},
		"mqtt_ca_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The url of the MQTT ca",
		},
		"mqtt_ca": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The MQTT certificat content",
		},
		"device_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of registered devices in the Hub",
		},
		"connected_device_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The current number of connected devices in the Hub",
		},
	}
}

func ResourceIotHubCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iotAPI, region, err := iotAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &iot.CreateHubRequest{
		Region:      region,
		Name:        types.ExpandOrGenerateString(d.Get("name"), "hub"),
		ProductPlan: iot.HubProductPlan(d.Get("product_plan").(string)),
	}

	if projectID, ok := d.GetOk("project_id"); ok {
		req.ProjectID = projectID.(string)
	}

	if disableEvents, ok := d.GetOk("disable_events"); ok {
		req.DisableEvents = scw.BoolPtr(disableEvents.(bool))
	}

	if eventsTopicPrefix, ok := d.GetOk("events_topic_prefix"); ok {
		req.EventsTopicPrefix = scw.StringPtr(eventsTopicPrefix.(string))
	}

	res, err := iotAPI.CreateHub(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, res.Region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitIotHub(ctx, iotAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	// Set user CA if needed. It cannot currently be added in the create hub request.
	if hubCA, ok := d.GetOk("hub_ca"); ok {
		_, err = iotAPI.SetHubCA(&iot.SetHubCARequest{
			Region:           region,
			HubID:            res.ID,
			CaCertPem:        hubCA.(string),
			ChallengeCertPem: d.Get("hub_ca_challenge").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Now user CA is set, set device auto provisioning if needed.
	if devProv, ok := d.GetOk("device_autoprovisioning"); ok {
		_, err = iotAPI.UpdateHub(&iot.UpdateHubRequest{
			EnableDeviceAutoProvisioning: scw.BoolPtr(devProv.(bool)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Disable hub if needed.
	// The only case we need to check is to eventually disable the hub (enabled by default)
	// We need to ensure the hub was fully enabled before because the hub cannot
	// be updated while enabling.
	if enabled := d.Get("enabled"); !enabled.(bool) {
		_, err = iotAPI.DisableHub(&iot.DisableHubRequest{
			Region: region,
			HubID:  res.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitIotHub(ctx, iotAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	MQTTUrl := computeIotHubCaURL(req.ProductPlan, region)
	_ = d.Set("mqtt_ca_url", MQTTUrl)

	return ResourceIotHubRead(ctx, d, m)
}

func ResourceIotHubRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iotAPI, region, hubID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := iotAPI.GetHub(&iot.GetHubRequest{
		Region: region,
		HubID:  hubID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", response.OrganizationID)
	_ = d.Set("project_id", response.ProjectID)
	_ = d.Set("name", response.Name)
	_ = d.Set("status", response.Status.String())
	_ = d.Set("product_plan", response.ProductPlan.String())
	_ = d.Set("endpoint", response.Endpoint)
	_ = d.Set("created_at", response.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", response.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("enabled", response.Enabled)
	_ = d.Set("device_count", int(response.DeviceCount))
	_ = d.Set("connected_device_count", int(response.ConnectedDeviceCount))
	_ = d.Set("disable_events", response.DisableEvents)
	_ = d.Set("events_topic_prefix", response.EventsTopicPrefix)
	_ = d.Set("device_auto_provisioning", response.EnableDeviceAutoProvisioning)
	_ = d.Set("mqtt_ca_url", computeIotHubCaURL(response.ProductPlan, region))
	mqttURL := d.Get("mqtt_ca_url")

	mqttCa, err := computeIotHubMQTTCa(ctx, fmt.Sprintf("%v", mqttURL), m)
	if err != nil {
		_ = diag.Diagnostic{
			Severity:      diag.Warning,
			AttributePath: cty.GetAttrPath("mqtt_ca"),
			Summary:       "Error while fetching the MQTT certificate.",
			Detail:        err.Error(),
		}
	}

	_ = d.Set("mqtt_ca", mqttCa)

	return nil
}

func ResourceIotHubUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iotAPI, region, hubID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Enable/Disable hub if needed
	////
	if d.HasChange("enabled") {
		newTargetStatus := d.Get("enabled").(bool)

		var err error
		if newTargetStatus {
			_, err = iotAPI.EnableHub(&iot.EnableHubRequest{
				Region: region,
				HubID:  hubID,
			}, scw.WithContext(ctx))
		} else {
			_, err = iotAPI.DisableHub(&iot.DisableHubRequest{
				Region: region,
				HubID:  hubID,
			}, scw.WithContext(ctx))
		}

		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitIotHub(ctx, iotAPI, region, hubID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////
	// Set the hub CA if changed
	////
	if d.HasChanges("hub_ca", "hub_ca_challenge") {
		_, err = iotAPI.SetHubCA(&iot.SetHubCARequest{
			Region:           region,
			HubID:            hubID,
			CaCertPem:        d.Get("hub_ca").(string),
			ChallengeCertPem: d.Get("hub_ca_challenge").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////
	// Construct UpdateHubRequest
	////
	updateRequest := &iot.UpdateHubRequest{
		Region: region,
		HubID:  hubID,
	}

	if d.HasChange("name") {
		updateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("product_plan") {
		updateRequest.ProductPlan = iot.HubProductPlan(d.Get("product_plan").(string))
	}

	if d.HasChange("disable_events") {
		updateRequest.DisableEvents = scw.BoolPtr(d.Get("disable_events").(bool))
	}

	if d.HasChange("events_topic_prefix") {
		updateRequest.EventsTopicPrefix = scw.StringPtr(d.Get("events_topic_prefix").(string))
	}

	if d.HasChange("device_auto_provisioning") {
		updateRequest.EnableDeviceAutoProvisioning = scw.BoolPtr(d.Get("device_auto_provisioning").(bool))
	}

	////
	// Apply Update
	////
	_, err = iotAPI.UpdateHub(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceIotHubRead(ctx, d, m)
}

func ResourceIotHubDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iotAPI, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = iotAPI.DeleteHub(&iot.DeleteHubRequest{
		Region: region,
		HubID:  id,
		// Don't force delete if devices. This avoids deleting a hub by mistake
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitIotHub(ctx, iotAPI, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
