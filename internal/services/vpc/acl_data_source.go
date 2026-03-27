package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceACL() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceACL().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "is_ipv6", "region")

	dsSchema["vpc_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The VPC ID to look up the ACL for",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceACLRead,
		Schema:      dsSchema,
	}
}

func DataSourceACLRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	vpcID := locality.ExpandID(d.Get("vpc_id").(string))

	regionalID := datasource.NewRegionalID(d.Get("vpc_id"), region)
	d.SetId(regionalID)

	acl, err := vpcAPI.GetACL(&vpc.GetACLRequest{
		VpcID:  vpcID,
		Region: region,
		IsIPv6: d.Get("is_ipv6").(bool),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("vpc_id", regional.NewIDString(region, vpcID))

	return setACLState(d, acl)
}
