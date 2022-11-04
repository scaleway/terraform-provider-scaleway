package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	sdkValidation "github.com/scaleway/scaleway-sdk-go/validation"
)

func resourceScalewayBaremetalServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayBaremetalServerCreate,
		ReadContext:   resourceScalewayBaremetalServerRead,
		UpdateContext: resourceScalewayBaremetalServerUpdate,
		DeleteContext: resourceScalewayBaremetalServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Create:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Update:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Delete:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the server",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname of the server",
			},
			"offer": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID or name of the server offer",
			},
			"offer_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the server offer",
			},
			"os": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The base image of the server",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"os_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The base image ID of the server",
			},
			"ssh_key_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validationUUID(),
				},
				Required: true,
				Description: `Array of SSH key IDs allowed to SSH to the server

**NOTE** : If you are attempting to update your SSH key IDs, it will induce the reinstall of your server. 
If this behaviour is wanted, please set 'reinstall_on_ssh_key_changes' argument to true.`,
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "User used for the installation.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password used for the installation.",
			},
			"service_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "User used for the service to install.",
			},
			"service_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password used for the service to install.",
			},
			"reinstall_on_config_changes": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If True, this boolean allows to reinstall the server on SSH key IDs, user or password changes",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Some description to associate to the server, max 255 characters",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Array of tags to associate with the server",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
			"ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the IP",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The version of the IP",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP address of the IP",
						},
						"reverse": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Reverse of the IP",
						},
					},
				},
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceScalewayBaremetalServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, zone, err := baremetalAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	offerID := expandZonedID(d.Get("offer"))
	if !sdkValidation.IsUUID(offerID.ID) {
		o, err := baremetalAPI.GetOfferByName(&baremetal.GetOfferByNameRequest{
			OfferName: offerID.ID,
			Zone:      zone,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		offerID = newZonedID(zone, o.ID)
	}
	if diags := validateInstallConfig(ctx, d, meta); len(diags) > 0 {
		return diags
	}

	server, err := baremetalAPI.CreateServer(&baremetal.CreateServerRequest{
		Zone:        zone,
		Name:        expandOrGenerateString(d.Get("name"), "bm"),
		ProjectID:   expandStringPtr(d.Get("project_id")),
		Description: d.Get("description").(string),
		OfferID:     offerID.ID,
		Tags:        expandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedID(server.Zone, server.ID).String())

	_, err = waitForBaremetalServer(ctx, baremetalAPI, zone, server.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = baremetalAPI.InstallServer(&baremetal.InstallServerRequest{
		Zone:            server.Zone,
		ServerID:        server.ID,
		OsID:            expandZonedID(d.Get("os")).ID,
		Hostname:        expandStringWithDefault(d.Get("hostname"), server.Name),
		SSHKeyIDs:       expandStrings(d.Get("ssh_key_ids")),
		User:            expandStringPtr(d.Get("user")),
		Password:        expandStringPtr(d.Get("password")),
		ServiceUser:     expandStringPtr(d.Get("service_user")),
		ServicePassword: expandStringPtr(d.Get("service_password")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBaremetalServerInstall(ctx, baremetalAPI, zone, server.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayBaremetalServerRead(ctx, d, meta)
}

func resourceScalewayBaremetalServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, zonedID, err := baremetalAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	server, err := baremetalAPI.GetServer(&baremetal.GetServerRequest{
		Zone:     zonedID.Zone,
		ServerID: zonedID.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	offer, err := baremetalAPI.GetOffer(&baremetal.GetOfferRequest{
		Zone:    server.Zone,
		OfferID: server.OfferID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", server.Name)
	_ = d.Set("zone", server.Zone.String())
	_ = d.Set("organization_id", server.OrganizationID)
	_ = d.Set("project_id", server.ProjectID)
	_ = d.Set("offer_id", newZonedID(server.Zone, offer.ID).String())
	_ = d.Set("tags", server.Tags)
	_ = d.Set("domain", server.Domain)
	_ = d.Set("ips", flattenBaremetalIPs(server.IPs))
	if server.Install != nil {
		_ = d.Set("os_id", newZonedID(server.Zone, server.Install.OsID).String())
		_ = d.Set("ssh_key_ids", server.Install.SSHKeyIDs)
		_ = d.Set("user", server.Install.User)
	}
	_ = d.Set("description", server.Description)

	return nil
}

func resourceScalewayBaremetalServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, zonedID, err := baremetalAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &baremetal.UpdateServerRequest{
		Zone:     zonedID.Zone,
		ServerID: zonedID.ID,
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = expandUpdatedStringPtr("name")
		hasChanged = true
	}

	if d.HasChange("description") {
		req.Description = expandUpdatedStringPtr("description")
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = expandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = baremetalAPI.UpdateServer(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	installReq := &baremetal.InstallServerRequest{
		Zone:            zonedID.Zone,
		ServerID:        zonedID.ID,
		Hostname:        expandStringWithDefault(d.Get("hostname"), d.Get("name").(string)),
		SSHKeyIDs:       expandStrings(d.Get("ssh_key_ids")),
		User:            expandStringPtr(d.Get("user")),
		Password:        expandStringPtr(d.Get("password")),
		ServiceUser:     expandStringPtr(d.Get("service_user")),
		ServicePassword: expandStringPtr(d.Get("service_password")),
	}

	if d.HasChange("os") {
		if diags := validateInstallConfig(ctx, d, meta); len(diags) > 0 {
			return diags
		}
		err = baremetalInstallServer(ctx, d, baremetalAPI, installReq)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForBaremetalServerInstall(ctx, baremetalAPI, zonedID.Zone, zonedID.ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	var diags diag.Diagnostics

	if d.HasChanges("ssh_key_ids", "user", "password", "reinstall_on_config_changes") {
		if !d.Get("reinstall_on_config_changes").(bool) && !d.HasChange("os") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Changes have been made on your config",
				Detail: "[WARN] This change induce the reinstall of your server. " +
					"If this behaviour is wanted, please set 'reinstall_on_config_changes' argument to true",
			})
		} else {
			if diags := validateInstallConfig(ctx, d, meta); len(diags) > 0 {
				return diags
			}
			err = baremetalInstallServer(ctx, d, baremetalAPI, installReq)
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForBaremetalServerInstall(ctx, baremetalAPI, zonedID.Zone, zonedID.ID, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return append(diags, resourceScalewayBaremetalServerRead(ctx, d, meta)...)
}

func resourceScalewayBaremetalServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, zonedID, err := baremetalAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = baremetalAPI.DeleteServer(&baremetal.DeleteServerRequest{
		Zone:     zonedID.Zone,
		ServerID: zonedID.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = waitForBaremetalServer(ctx, baremetalAPI, zonedID.Zone, zonedID.ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

func baremetalInstallAttributeMissing(field *baremetal.OSOSField, d *schema.ResourceData, attribute string) bool {
	if field != nil && field.Required && field.DefaultValue == nil {
		if _, attributeExists := d.GetOk(attribute); !attributeExists {
			return true
		}
	}
	return false
}

// validateInstallConfig validates that schema contains attribute required for OS install
func validateInstallConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, zone, err := baremetalAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	os, err := baremetalAPI.GetOS(&baremetal.GetOSRequest{
		Zone: zone,
		OsID: expandID(d.Get("os")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	diags := diag.Diagnostics(nil)
	installAttributes := []struct {
		Attribute string
		Field     *baremetal.OSOSField
	}{
		{
			"user",
			os.User,
		},
		{
			"password",
			os.Password,
		},
		{
			"service_user",
			os.ServiceUser,
		},
		{
			"service_password",
			os.ServicePassword,
		},
	}
	for _, installAttr := range installAttributes {
		if baremetalInstallAttributeMissing(installAttr.Field, d, installAttr.Attribute) {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("%s attribute is required", installAttr.Attribute),
				Detail:        fmt.Sprintf("%s is required for this os", installAttr.Attribute),
				AttributePath: cty.GetAttrPath(installAttr.Attribute),
			})
		}
	}
	return diags
}
