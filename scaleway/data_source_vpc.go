package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPC() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPC().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "is_default", "region")

	dsSchema["name"].ConflictsWith = []string{"vpc_id"}
	dsSchema["vpc_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the VPC",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["organization_id"] = organizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The project ID the resource is associated to",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCRead,
	}
}

func dataSourceScalewayVPCRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var vpcID interface{}
	var ok bool

	if d.Get("is_default").(bool) {
		request := &vpc.ListVPCsRequest{
			IsDefault: expandBoolPtr(d.Get("is_default").(bool)),
			Region:    region,
		}

		res, err := vpcAPI.ListVPCs(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		vpcID = newRegionalIDString(region, res.Vpcs[0].ID)
	} else {
		vpcID, ok = d.GetOk("vpc_id")
		if !ok {
			request := &vpc.ListVPCsRequest{
				Name:           expandStringPtr(d.Get("name").(string)),
				Region:         region,
				ProjectID:      expandStringPtr(d.Get("project_id")),
				OrganizationID: expandStringPtr(d.Get("organization_id")),
			}

			res, err := vpcAPI.ListVPCs(request, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			for _, v := range res.Vpcs {
				if v.Name == d.Get("name").(string) {
					if vpcID != "" {
						return diag.FromErr(fmt.Errorf("more than 1 VPC found with the same name %s", d.Get("name")))
					}
					vpcID = newRegionalIDString(region, v.ID)
				}
			}
			if res.TotalCount == 0 {
				return diag.FromErr(fmt.Errorf("no VPC found with the name %s", d.Get("name")))
			}
		}
	}

	regionalID := datasourceNewRegionalizedID(vpcID, region)
	d.SetId(regionalID)
	err = d.Set("vpc_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayVPCRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read VPC")...)
	}

	if d.Id() == "" {
		return diag.Errorf("VPC (%s) not found", regionalID)
	}

	return nil
}
