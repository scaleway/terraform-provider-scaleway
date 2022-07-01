package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayFlexibleIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFlexibleIPCreate,
		ReadContext:   resourceScalewayFlexibleIPRead,
		UpdateContext: resourceScalewayFlexibleIPUpdate,
		DeleteContext: resourceScalewayFlexibleIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultFlexibleIPTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the flexible IP",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv4 address of the flexible IP",
			},
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The reverse DNS for this flexible IP",
			},
			"server_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The server associated with this flexible IP",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MAC address of the server associated with this flexible IP",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the flexible IP",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the Flexible IP",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the Flexible IP",
			},
		},
	}
}

func resourceScalewayFlexibleIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := fipAPI.CreateFlexibleIP(&flexibleip.CreateFlexibleIPRequest{
		Zone:        zone,
		ProjectID:   d.Get("project_id").(string),
		Description: d.Get("description").(string),
		Tags:        expandStrings(d.Get("tags")),
		ServerID:    expandStringPtr(expandID(d.Get("server_id"))),
		Reverse:     expandStringPtr(d.Get("reverse")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, flexibleIP.ID))

	_, err = waitFlexibleIP(ctx, fipAPI, zone, flexibleIP.ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayFlexibleIPRead(ctx, d, meta)
}

func resourceScalewayFlexibleIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fipAPI, zone, ID, err := fipAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// verify resource is ready
	_, err = waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := fipAPI.GetFlexibleIP(&flexibleip.GetFlexibleIPRequest{
		Zone:  zone,
		FipID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		// We check for 403 because flexible API returns 403 for a deleted IP
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("ip_address", flexibleIP.IPAddress.String())
	_ = d.Set("zone", zone) // TODO use zone field from flexibleIP when available
	_ = d.Set("organization_id", flexibleIP.OrganizationID)
	_ = d.Set("project_id", flexibleIP.ProjectID)
	_ = d.Set("reverse", flexibleIP.Reverse)
	_ = d.Set("created_at", flexibleIP.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", flexibleIP.UpdatedAt.Format(time.RFC3339))

	if flexibleIP.ServerID != nil {
		_ = d.Set("server_id", newZonedIDString(zone, *flexibleIP.ServerID))
	} else {
		_ = d.Set("server_id", "")
	}

	return nil
}

func resourceScalewayFlexibleIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fipAPI, zone, ID, err := fipAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	updateRequest := &flexibleip.UpdateFlexibleIPRequest{
		Zone:  zone,
		FipID: flexibleIP.ID,
	}

	if d.HasChanges("reverse") {
		updateRequest.Reverse = expandStringPtr(d.Get("reverse"))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = scw.StringsPtr(expandStrings(d.Get("tags")))
	}

	if d.HasChange("description") {
		updateRequest.Description = expandStringPtr(d.Get("description"))
	}

	_, err = fipAPI.UpdateFlexibleIP(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("server_id") {
		if _, serverIDExists := d.GetOk("server_id"); !serverIDExists {
			_, err = fipAPI.DetachFlexibleIP(&flexibleip.DetachFlexibleIPRequest{
				Zone:    zone,
				FipsIDs: []string{ID},
			})
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			_, err = fipAPI.AttachFlexibleIP(&flexibleip.AttachFlexibleIPRequest{
				Zone:     zone,
				FipsIDs:  []string{ID},
				ServerID: expandID(d.Get("server_id")),
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	_, err = waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayFlexibleIPRead(ctx, d, meta)
}

func resourceScalewayFlexibleIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fipAPI, zone, ID, err := fipAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = fipAPI.DeleteFlexibleIP(&flexibleip.DeleteFlexibleIPRequest{
		FipID: flexibleIP.ID,
		Zone:  zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) && !is403Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitFlexibleIP(ctx, fipAPI, zone, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) && !is403Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
