package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPCPrivateNetwork() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPrivateNetwork().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "project_id")

	dsSchema["name"].ConflictsWith = []string{"private_network_id"}
	dsSchema["vpc_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the vpc to which the private network belongs to",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"private_network_id"},
	}
	dsSchema["private_network_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the private network",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name", "vpc_id"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPrivateNetworkRead,
	}
}

func dataSourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok {
		pnName := d.Get("name").(string)
		res, err := vpcAPI.ListPrivateNetworks(
			&vpc.ListPrivateNetworksRequest{
				Name:      expandStringPtr(pnName),
				Region:    region,
				ProjectID: expandStringPtr(d.Get("project_id")),
				VpcID:     expandStringPtr(expandID(d.Get("vpc_id"))),
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPN, err := findExact(
			res.PrivateNetworks,
			func(s *vpc.PrivateNetwork) bool { return s.Name == pnName },
			pnName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		privateNetworkID = foundPN.ID
	}

	regionalID := datasourceNewRegionalID(privateNetworkID, region)
	d.SetId(regionalID)
	_ = d.Set("private_network_id", regionalID)
	diags := resourceScalewayVPCPrivateNetworkRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read private network state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("private network (%s) not found", regionalID)
	}

	return nil
}
