package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayVPCPrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPrivateNetworkCreate,
		ReadContext:   resourceScalewayVPCPrivateNetworkRead,
		UpdateContext: resourceScalewayVPCPrivateNetworkUpdate,
		DeleteContext: resourceScalewayVPCPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the private network",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with private network",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": projectIDSchema(),
			"zone":       zoneSchema(),
		},
	}
}

func resourceScalewayVPCPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	vpcAPI, zone, err := vpcAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.CreatePrivateNetwork(&vpc.CreatePrivateNetworkRequest{
		Name:      expandOrGenerateString(d.Get("name"), "pn"),
		Tags:      expandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Zone:      zone,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	return resourceScalewayVPCPrivateNetworkRead(ctx, d, m)
}

func resourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pn, err := vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("id", newZonedIDString(zone, pn.ID))
	_ = d.Set("name", pn.Name)
	_ = d.Set("organization_id", pn.OrganizationID)
	_ = d.Set("project_id", pn.ProjectID)
	_ = d.Set("created_at", pn.CreatedAt.String())
	_ = d.Set("updated_at", pn.UpdatedAt.String())
	_ = d.Set("zone", zone)
	_ = d.Set("tags", pn.Tags)

	return nil
}

func resourceScalewayVPCPrivateNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &vpc.UpdatePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
	}

	if d.HasChange("name") {
		updateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		tags := expandStrings(d.Get("tags"))
		updateRequest.Tags = scw.StringsPtr(tags)
	}

	_, err = vpcAPI.UpdatePrivateNetwork(updateRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPrivateNetworkRead(ctx, d, m)
}

func resourceScalewayVPCPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcAPI.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
	})

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
