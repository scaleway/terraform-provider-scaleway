package webhosting

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceWebhosting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhostingCreate,
		ReadContext:   resourceWebhostingRead,
		UpdateContext: resourceWebhostingUpdate,
		DeleteContext: resourceHostingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultHostingTimeout),
			Read:    schema.DefaultTimeout(defaultHostingTimeout),
			Delete:  schema.DefaultTimeout(defaultHostingTimeout),
			Default: schema.DefaultTimeout(defaultHostingTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"offer_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "The ID of the selected offer for the hosting",
			},
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.IsEmail(),
				Description:      "Contact email of the client for the hosting",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain name of the hosting",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true,
				Description: "The tags of the hosting",
			},
			"option_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the selected options for the hosting",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IsDate and time of hosting's creation (RFC 3339 format)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IsDate and time of hosting's last update (RFC 3339 format)",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The hosting status",
			},
			"platform_hostname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hostname of the host platform",
			},
			"platform_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of the host platform",
			},
			"offer_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the active offer",
			},
			"options": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Active options of the hosting",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS status of the hosting",
			},
			"cpanel_urls": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "URL to connect to cPanel Dashboard and to Webmail interface",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dashboard": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"webmail": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Main hosting cPanel username",
			},
			"region":          regional.Schema(),
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
		},
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
			if diff.HasChange("tags") {
				oldTagsInterface, newTagsInterface := diff.GetChange("tags")
				oldTags := types.ExpandStrings(oldTagsInterface)
				newTags := types.ExpandStrings(newTagsInterface)
				// If the 'internal' tag has been added, remove it from the diff
				if types.SliceContainsString(oldTags, "internal") && !types.SliceContainsString(newTags, "internal") {
					err := diff.SetNew("tags", oldTags)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}
}

func resourceWebhostingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, offerID, err := regional.ParseID(d.Get("offer_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	hostingCreateRequest := &webhosting.CreateHostingRequest{
		Region:    region,
		OfferID:   offerID,
		ProjectID: d.Get("project_id").(string),
		Domain:    d.Get("domain").(string),
		OptionIDs: types.ExpandStrings(d.Get("option_ids")),
	}

	rawTags, tagsExist := d.GetOk("tags")
	if tagsExist {
		hostingCreateRequest.Tags = types.ExpandStrings(rawTags)
	}

	rawOptionIDs, rawOptionIDsExist := d.GetOk("option_ids")
	if rawOptionIDsExist {
		hostingCreateRequest.OptionIDs = types.ExpandStrings(rawOptionIDs)
	}

	rawEmail, emailExist := d.GetOk("email")
	if emailExist {
		hostingCreateRequest.Email = types.ExpandStringPtr(rawEmail)
	}

	hostingResponse, err := api.CreateHosting(hostingCreateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, hostingResponse.ID))

	_, err = waitForHosting(ctx, api, region, hostingResponse.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWebhostingRead(ctx, d, m)
}

func resourceWebhostingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	webhostingResponse, err := waitForHosting(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("tags", webhostingResponse.Tags)
	_ = d.Set("offer_id", regional.NewIDString(region, webhostingResponse.OfferID))
	_ = d.Set("domain", webhostingResponse.Domain)
	_ = d.Set("created_at", types.FlattenTime(webhostingResponse.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(webhostingResponse.UpdatedAt))
	_ = d.Set("status", webhostingResponse.Status.String())
	_ = d.Set("platform_hostname", webhostingResponse.PlatformHostname)
	_ = d.Set("platform_number", types.FlattenInt32Ptr(webhostingResponse.PlatformNumber))
	_ = d.Set("options", flattenHostingOptions(webhostingResponse.Options))
	_ = d.Set("offer_name", webhostingResponse.OfferName)
	_ = d.Set("dns_status", webhostingResponse.DNSStatus.String())
	_ = d.Set("cpanel_urls", flattenHostingCpanelUrls(webhostingResponse.CpanelURLs))
	_ = d.Set("username", webhostingResponse.Username)
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", webhostingResponse.OrganizationID)
	_ = d.Set("project_id", webhostingResponse.ProjectID)

	return nil
}

func resourceWebhostingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := waitForHosting(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &webhosting.UpdateHostingRequest{
		Region:    region,
		HostingID: res.ID,
	}

	hasChanged := false

	if d.HasChange("option_ids") {
		updateRequest.OptionIDs = types.ExpandUpdatedStringsPtr(d.Get("option_ids"))
		hasChanged = true
	}

	if d.HasChange("offer_id") {
		_, offerID, err := regional.ParseID(d.Get("offer_id").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		updateRequest.OfferID = types.ExpandUpdatedStringPtr(offerID)
		hasChanged = true
	}

	if d.HasChange("email") {
		updateRequest.Email = types.ExpandUpdatedStringPtr(d.Get("email"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateHosting(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWebhostingRead(ctx, d, m)
}

func resourceHostingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForHosting(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return nil
	}

	_, err = api.DeleteHosting(&webhosting.DeleteHostingRequest{
		Region:    region,
		HostingID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForHosting(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return nil
	}

	return nil
}
