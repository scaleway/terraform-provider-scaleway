package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
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

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPrivateNetworkRead,
	}
}

func dataSourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	vpcAPI, zone, err := vpcAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok {
		res, err := vpcAPI.ListPrivateNetworks(
			&vpc.ListPrivateNetworksRequest{
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
	return resourceScalewayVPCPrivateNetworkRead(ctx, d, m)
}
