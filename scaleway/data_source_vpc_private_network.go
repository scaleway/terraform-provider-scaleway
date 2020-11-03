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
		ReadContext: dataSourceScalewayVPCPrivateNetworkRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, err := vpcAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok { // Get private networks by zone and name.
		res, err := vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
			Zone:      zone,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.PrivateNetworks) == 0 {
			return diag.FromErr(fmt.Errorf("no private networks found with the name %s", d.Get("name")))
		}
		if len(res.PrivateNetworks) > 1 {
			return diag.FromErr(fmt.Errorf("%d private networks found with the same name %s", len(res.PrivateNetworks), d.Get("name")))
		}
		privateNetworkID = res.PrivateNetworks[0].ID
	}

	zonedID := datasourceNewZonedID(privateNetworkID, zone)
	d.SetId(zonedID)
	err = d.Set("private_network_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayVPCPrivateNetworkRead(ctx, d, meta)
}
