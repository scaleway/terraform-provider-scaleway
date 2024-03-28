package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourcePrivateNetwork() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePrivateNetwork().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "project_id")

	dsSchema["name"].ConflictsWith = []string{"private_network_id"}
	dsSchema["vpc_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the vpc to which the private network belongs to",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"private_network_id"},
	}
	dsSchema["private_network_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the private network",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name", "vpc_id"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCPrivateNetworkRead,
	}
}

func DataSourceVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkID, ok := d.GetOk("private_network_id")
	if !ok {
		pnName := d.Get("name").(string)
		res, err := vpcAPI.ListPrivateNetworks(
			&vpc.ListPrivateNetworksRequest{
				Name:      types.ExpandStringPtr(pnName),
				Region:    region,
				ProjectID: types.ExpandStringPtr(d.Get("project_id")),
				VpcID:     types.ExpandStringPtr(locality.ExpandID(d.Get("vpc_id"))),
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPN, err := datasource.FindExact(
			res.PrivateNetworks,
			func(s *vpc.PrivateNetwork) bool { return s.Name == pnName },
			pnName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		privateNetworkID = foundPN.ID
	}

	regionalID := datasource.NewRegionalID(privateNetworkID, region)
	d.SetId(regionalID)
	_ = d.Set("private_network_id", regionalID)
	diags := ResourceVPCPrivateNetworkRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read private network state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("private network (%s) not found", regionalID)
	}

	return nil
}
