package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPCPrivateNetwork() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPrivateNetwork().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"private_network_id"}
	dsSchema["private_network_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the private network",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["is_regional"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Whether this is a regional or zonal private network",
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPrivateNetworkRead,
	}
}

func dataSourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("is_regional").(bool) {
		return dataSourceScalewayVPCPrivateNetworkRegionalRead(ctx, d, meta)
	}
	return dataSourceScalewayVPCPrivateNetworkZonalRead(ctx, d, meta)
}

func dataSourceScalewayVPCPrivateNetworkZonalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, err := vpcAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok {
		res, err := vpcAPI.ListPrivateNetworks(
			&v1.ListPrivateNetworksRequest{
				Name: expandStringPtr(d.Get("name").(string)),
				Zone: zone,
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if res.TotalCount == 0 {
			return diag.FromErr(
				fmt.Errorf(
					"no private network found with the name %s",
					d.Get("name"),
				),
			)
		}
		if res.TotalCount > 1 {
			return diag.FromErr(
				fmt.Errorf(
					"%d private networks found with the name %s",
					res.TotalCount,
					d.Get("name"),
				),
			)
		}
		privateNetworkID = res.PrivateNetworks[0].ID
	}

	zonedID := datasourceNewZonedID(privateNetworkID, zone)
	d.SetId(zonedID)
	_ = d.Set("private_network_id", zonedID)
	return resourceScalewayVPCPrivateNetworkZonalRead(ctx, d, meta)
}

func dataSourceScalewayVPCPrivateNetworkRegionalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok {
		res, err := vpcAPI.ListPrivateNetworks(
			&v2.ListPrivateNetworksRequest{
				Name:   expandStringPtr(d.Get("name").(string)),
				Region: region,
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if res.TotalCount == 0 {
			return diag.FromErr(
				fmt.Errorf(
					"no private network found with the name %s",
					d.Get("name"),
				),
			)
		}
		if res.TotalCount > 1 {
			return diag.FromErr(
				fmt.Errorf(
					"%d private networks found with the name %s",
					res.TotalCount,
					d.Get("name"),
				),
			)
		}
		privateNetworkID = res.PrivateNetworks[0].ID
	}

	regionalID := datasourceNewRegionalizedID(privateNetworkID, region)
	d.SetId(regionalID)
	_ = d.Set("private_network_id", regionalID)
	return resourceScalewayVPCPrivateNetworkRegionalRead(ctx, d, meta)
}
