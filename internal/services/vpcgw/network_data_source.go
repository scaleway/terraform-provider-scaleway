package vpcgw

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceNetwork() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceNetwork().SchemaFunc())

	searchFields := []string{
		"gateway_id",
		"private_network_id",
		"enable_masquerade",
		"dhcp_id",
	}
	datasource.AddOptionalFieldsToSchema(dsSchema, searchFields...)

	dsSchema["gateway_network_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the gateway network",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    searchFields,
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCGatewayNetworkRead,
	}
}

func DataSourceVPCGatewayNetworkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetworkID, ok := d.GetOk("gateway_network_id")
	if !ok {
		res, err := api.ListGatewayNetworks(&vpcgw.ListGatewayNetworksRequest{
			GatewayIDs:        []string{locality.ExpandID(d.Get("gateway_id").(string))},
			PrivateNetworkIDs: []string{locality.ExpandID(d.Get("private_network_id"))},
			MasqueradeEnabled: types.ExpandBoolPtr(types.GetBool(d, "enable_masquerade")),
			Zone:              zone,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if res.TotalCount == 0 {
			return diag.FromErr(errors.New("no gateway network found with the filters"))
		}

		if res.TotalCount > 1 {
			return diag.FromErr(fmt.Errorf("%d gateway networks found with filters", res.TotalCount))
		}

		gatewayNetworkID = res.GatewayNetworks[0].ID
	}

	zonedID := datasource.NewZonedID(gatewayNetworkID, zone)
	d.SetId(zonedID)
	_ = d.Set("gateway_network_id", zonedID)

	gatewayNetwork, err := api.GetGatewayNetwork(&vpcgw.GetGatewayNetworkRequest{
		GatewayNetworkID: locality.ExpandID(gatewayNetworkID),
		Zone:             zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if gatewayNetwork.PrivateNetworkID != "" {
		diags = setPrivateIPs(ctx, d, api, gatewayNetwork, m)
	}

	diags = append(diags, setGatewayNetworkState(d, gatewayNetwork)...)

	return diags
}
