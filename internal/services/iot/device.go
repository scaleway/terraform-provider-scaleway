package iot

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	iotPolicySuffix = ".policy"
	iotTopicsSuffix = ".topics"
)

func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceIotDeviceCreate,
		ReadContext:   ResourceIotDeviceRead,
		UpdateContext: ResourceIotDeviceUpdate,
		DeleteContext: ResourceIotDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"hub_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The ID of the hub on which this device will be created",
				DiffSuppressFunc: dsf.Locality,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the device",
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the device",
			},
			"allow_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allow plain and server-authenticated SSL connections in addition to mutually-authenticated ones",
			},
			"allow_multiple_connections": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allow multiple connections",
			},
			"message_filters": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Rules to authorize or deny the device to publish/subscribe to specific topics",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"publish": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Rule to restrict topics the device can publish to",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy": {
										Type:             schema.TypeString,
										Optional:         true,
										Description:      "Publish message filter policy",
										Default:          iot.DeviceMessageFiltersRulePolicyReject.String(),
										RequiredWith:     []string{"message_filters.0.publish.0.topics"},
										ValidateDiagFunc: verify.ValidateEnum[iot.DeviceMessageFiltersRulePolicy](),
									},
									"topics": {
										Type:         schema.TypeList,
										Optional:     true,
										Description:  "List of topics in the set",
										RequiredWith: []string{"message_filters.0.publish.0.policy"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"subscribe": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Rule to restrict topics the device can subscribe to",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy": {
										Type:             schema.TypeString,
										Optional:         true,
										Description:      "Subscribe message filter policy",
										Default:          iot.DeviceMessageFiltersRulePolicyReject.String(),
										RequiredWith:     []string{"message_filters.0.subscribe.0.topics"},
										ValidateDiagFunc: verify.ValidateEnum[iot.DeviceMessageFiltersRulePolicy](),
									},
									"topics": {
										Type:         schema.TypeList,
										Optional:     true,
										Description:  "List of topics in the set",
										RequiredWith: []string{"message_filters.0.subscribe.0.policy"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			// Provided or computed elements
			"certificate": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Certificate section of the device",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"crt": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Sensitive:   true,
							Description: "X509 PEM encoded certificate of the device",
						},
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "X509 PEM encoded key of the device",
						},
					},
				},
			},
			// Computed elements
			"region": regional.Schema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the device",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the device",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the device",
			},
			"last_activity_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of last MQTT activity of the device",
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The MQTT connection status of the device",
			},
		},
		CustomizeDiff: cdf.LocalityCheck("hub_id"),
	}
}

func ResourceIotDeviceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, err := iotAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &iot.CreateDeviceRequest{
		Region: region,
		HubID:  locality.ExpandID(d.Get("hub_id")),
		Name:   types.ExpandOrGenerateString(d.Get("name"), "device"),
	}

	if allowInsecure, ok := d.GetOk("allow_insecure"); ok {
		req.AllowInsecure = allowInsecure.(bool)
	}

	if allowMultipleConn, ok := d.GetOk("allow_multiple_connections"); ok {
		req.AllowMultipleConnections = allowMultipleConn.(bool)
	}

	if description, ok := d.GetOk("description"); ok {
		req.Description = scw.StringPtr(description.(string))
	}

	if _, ok := d.GetOk("message_filters"); ok {
		mf := &iot.DeviceMessageFilters{}

		fqfn := "message_filters.0"
		if _, ok := d.GetOk(fqfn + ".publish"); ok {
			fqfnS := fqfn + ".publish.0"
			mfSet := iot.DeviceMessageFiltersRule{}

			if policy, ok := d.GetOk(fqfnS + iotPolicySuffix); ok {
				mfSet.Policy = iot.DeviceMessageFiltersRulePolicy(policy.(string))
			}

			if topics, ok := d.GetOk(fqfnS + iotTopicsSuffix); ok {
				mfSet.Topics = scw.StringsPtr(types.ExpandStringsOrEmpty(topics))
			}

			mf.Publish = &mfSet
		}

		if _, ok := d.GetOk(fqfn + ".subscribe"); ok {
			fqfnP := fqfn + ".subscribe.0"
			mfSet := iot.DeviceMessageFiltersRule{}

			if policy, ok := d.GetOk(fqfnP + iotPolicySuffix); ok {
				mfSet.Policy = iot.DeviceMessageFiltersRulePolicy(policy.(string))
			}

			if topics, ok := d.GetOk(fqfnP + iotTopicsSuffix); ok {
				mfSet.Topics = scw.StringsPtr(types.ExpandStringsOrEmpty(topics))
			}

			mf.Subscribe = &mfSet
		}

		req.MessageFilters = mf
	}

	res, err := iotAPI.CreateDevice(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.Device.ID))

	// If user certificate is provided.
	if devCrt, ok := d.GetOk("certificate.0.crt"); ok {
		// Set user certificate to device.
		// It cannot currently be added in the create device request.
		_, err := iotAPI.SetDeviceCertificate(&iot.SetDeviceCertificateRequest{
			Region:         region,
			DeviceID:       res.Device.ID,
			CertificatePem: devCrt.(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Update certificate and key as they cannot be retrieved later.
		cert := map[string]interface{}{
			"crt": res.Certificate.Crt,
			"key": res.Certificate.Key,
		}
		_ = d.Set("certificate", []map[string]interface{}{cert})
	}

	return ResourceIotDeviceRead(ctx, d, m)
}

func ResourceIotDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, deviceID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	device, err := iotAPI.GetDevice(&iot.GetDeviceRequest{
		Region:   region,
		DeviceID: deviceID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", device.Name)
	_ = d.Set("status", device.Status)
	_ = d.Set("hub_id", regional.NewID(region, device.HubID).String())
	_ = d.Set("created_at", device.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", device.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("last_activity_at", device.LastActivityAt.Format(time.RFC3339))
	_ = d.Set("allow_insecure", device.AllowInsecure)
	_ = d.Set("allow_multiple_connections", device.AllowMultipleConnections)
	_ = d.Set("is_connected", device.IsConnected)
	_ = d.Set("description", device.Description)

	mf := map[string]interface{}{}
	mfHasNonDefaultChange := false

	// We need to set the message filters only in case when we already set a value or we got non default value

	// In case of already set value
	if _, ok := d.GetOk("message_filters.0"); ok {
		mfHasNonDefaultChange = true
	}
	// In case of non default change
	if device.MessageFilters.Publish.Policy != iot.DeviceMessageFiltersRulePolicyReject ||
		(device.MessageFilters.Publish.Topics != nil && len(*device.MessageFilters.Publish.Topics) != 0) {
		mfHasNonDefaultChange = true
	}

	if device.MessageFilters.Subscribe.Policy != iot.DeviceMessageFiltersRulePolicyReject ||
		(device.MessageFilters.Subscribe.Topics != nil && len(*device.MessageFilters.Subscribe.Topics) != 0) {
		mfHasNonDefaultChange = true
	}

	if mfHasNonDefaultChange {
		mf["publish"] = []map[string]interface{}{{
			"policy": device.MessageFilters.Publish.Policy,
			"topics": *device.MessageFilters.Publish.Topics,
		}}

		mf["subscribe"] = []map[string]interface{}{{
			"policy": device.MessageFilters.Subscribe.Policy,
			"topics": *device.MessageFilters.Subscribe.Topics,
		}}
	}

	if mfHasNonDefaultChange {
		_ = d.Set("message_filters", []map[string]interface{}{mf})
	}

	// Read Device certificate
	// As we cannot read the key, we get back it from state and do not change it.
	if devCrtKey, ok := d.GetOk("certificate.0.key"); ok {
		devCrt, err := iotAPI.GetDeviceCertificate(&iot.GetDeviceCertificateRequest{
			Region:   region,
			DeviceID: deviceID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		// Set device certificate.
		cert := map[string]interface{}{
			"crt": devCrt.CertificatePem,
			"key": devCrtKey.(string),
		}
		_ = d.Set("certificate", []map[string]interface{}{cert})
	}

	return nil
}

func ResourceIotDeviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, deviceID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &iot.UpdateDeviceRequest{
		Region:   region,
		DeviceID: deviceID,
	}

	if d.HasChange("allow_insecure") {
		updateRequest.AllowInsecure = scw.BoolPtr(d.Get("allow_insecure").(bool))
	}

	if d.HasChange("message_filters") {
		fqfn := "message_filters.0"
		mf := &iot.DeviceMessageFilters{}
		updateRequest.MessageFilters = mf

		if d.HasChange(fqfn + ".publish") {
			fqfnS := fqfn + ".publish.0"
			mfSet := iot.DeviceMessageFiltersRule{}
			mf.Publish = &mfSet

			mfSet.Policy = iot.DeviceMessageFiltersRulePolicy(
				d.Get(fqfnS + iotPolicySuffix).(string))

			mfSet.Topics = scw.StringsPtr(
				types.ExpandStringsOrEmpty(d.Get(fqfnS + iotTopicsSuffix)))
		}

		if d.HasChange(fqfn + ".subscribe") {
			fqfnP := fqfn + ".subscribe.0"
			mfSet := iot.DeviceMessageFiltersRule{}
			mf.Subscribe = &mfSet

			mfSet.Policy = iot.DeviceMessageFiltersRulePolicy(
				d.Get(fqfnP + iotPolicySuffix).(string))

			mfSet.Topics = scw.StringsPtr(
				types.ExpandStringsOrEmpty(d.Get(fqfnP + iotTopicsSuffix)))
		}
	}

	if d.HasChange("hub_id") {
		updateRequest.HubID = scw.StringPtr(d.Get("hub_id").(string))
	}

	if d.HasChange("allow_multiple_connections") {
		updateRequest.AllowMultipleConnections = scw.BoolPtr(d.Get("allow_multiple_connections").(bool))
	}

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	_, err = iotAPI.UpdateDevice(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the device certificate if changed
	if d.HasChange("certificate.0.crt") {
		_, err := iotAPI.SetDeviceCertificate(&iot.SetDeviceCertificateRequest{
			Region:         region,
			DeviceID:       deviceID,
			CertificatePem: d.Get("certificate.0.crt").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceIotDeviceRead(ctx, d, m)
}

func ResourceIotDeviceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, deviceID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = iotAPI.DeleteDevice(&iot.DeleteDeviceRequest{
		Region:   region,
		DeviceID: deviceID,
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
