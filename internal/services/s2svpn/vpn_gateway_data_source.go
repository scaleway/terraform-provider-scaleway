package s2svpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2svpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceVPNGateway() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceVPNGateway().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"vpn_gateway_id"}
	dsSchema["vpn_gateway_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the VPN gateway",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceS2SVPNGatewayRead,
	}
}

func DataSourceS2SVPNGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	vpnGatewayID, ok := d.GetOk("vpn_gateway_id")
	if !ok {
		gatewayName := d.Get("name").(string)

		res, err := api.ListVpnGateways(&s2svpn.ListVpnGatewaysRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(gatewayName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundGateway, err := datasource.FindExact(
			res.Gateways,
			func(s *s2svpn.VpnGateway) bool { return s.Name == gatewayName },
			gatewayName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		vpnGatewayID = foundGateway.ID
	}

	regionalID := datasource.NewRegionalID(vpnGatewayID, region)
	d.SetId(regionalID)
	_ = d.Set("vpn_gateway_id", regionalID)

	diags := ResourceVPNGatewayRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read VPN gateway state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("VPN gateway (%s) not found", regionalID)
	}

	return nil
}
