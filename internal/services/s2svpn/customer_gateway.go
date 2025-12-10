package s2svpn

import (
	"context"
	"net"
	_ "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCustomerGatewayCreate,
		ReadContext:   ResourceCustomerGatewayRead,
		UpdateContext: ResourceCustomerGatewayUpdate,
		DeleteContext: ResourceCustomerGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the customer gateway",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of tags to apply to the customer gateway",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ipv4_public": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The public IPv4 address of the customer gateway",
			},
			"ipv6_public": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The public IPv6 address of the customer gateway",
			},
			"asn": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The AS Number of the customer gateway",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the TLS stage",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the TLS stage",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID of the Project",
			},
		},
	}
}

func ResourceCustomerGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &s2s_vpn.CreateCustomerGatewayRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      types.ExpandOrGenerateString(d.Get("name").(string), "connection"),
		Tags:      types.ExpandStrings(d.Get("tags")),
		Asn:       uint32(d.Get("asn").(int)),
	}

	if d.Get("ipv4_public").(string) != "" {
		req.IPv4Public = scw.IPPtr(net.ParseIP(d.Get("ipv4_public").(string)))
	}

	if d.Get("ipv6_public").(string) != "" {
		req.IPv6Public = scw.IPPtr(net.ParseIP(d.Get("ipv6_public").(string)))
	}

	res, err := api.CreateCustomerGateway(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceCustomerGatewayRead(ctx, d, m)
}

func ResourceCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gateway, err := api.GetCustomerGateway(&s2s_vpn.GetCustomerGatewayRequest{
		GatewayID: id,
		Region:    region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", gateway.Name)
	_ = d.Set("project_id", gateway.ProjectID)
	_ = d.Set("organization_id", gateway.OrganizationID)
	_ = d.Set("tags", gateway.Tags)
	_ = d.Set("created_at", types.FlattenTime(gateway.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(gateway.UpdatedAt))
	_ = d.Set("ipv4_public", types.FlattenIPPtr(gateway.PublicIPv4))
	_ = d.Set("ipv6_public", types.FlattenIPPtr(gateway.PublicIPv6))
	_ = d.Set("asn", int(gateway.Asn))

	return nil
}

func ResourceCustomerGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	req := &s2s_vpn.UpdateCustomerGatewayRequest{
		Region:    region,
		GatewayID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("ipv4_public") {
		req.IPv4Public = scw.IPPtr(net.ParseIP(d.Get("ipv4_public").(string)))
		hasChanged = true
	}

	if d.HasChange("ipv6_public") {
		req.IPv6Public = scw.IPPtr(net.ParseIP(d.Get("ipv6_public").(string)))
		hasChanged = true
	}

	if d.HasChange("asn") {
		req.Asn = types.ExpandUint32Ptr(d.Get("asn"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateCustomerGateway(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceCustomerGatewayRead(ctx, d, m)
}

func ResourceCustomerGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteCustomerGateway(&s2s_vpn.DeleteCustomerGatewayRequest{
		Region:    region,
		GatewayID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
