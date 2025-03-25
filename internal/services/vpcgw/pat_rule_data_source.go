package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourcePATRule() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePATRule().Schema)

	dsSchema["pat_rule_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The ID of the public gateway PAT rule",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "zone")

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCPublicGatewayPATRuleRead,
	}
}

func DataSourceVPCPublicGatewayPATRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	patRuleIDRaw := d.Get("pat_rule_id")

	zonedID := datasource.NewZonedID(patRuleIDRaw, zone)
	d.SetId(zonedID)
	_ = d.Set("pat_rule_id", zonedID)

	// check if pat rule exist
	_, err = api.GetPatRule(&vpcgw.GetPatRuleRequest{
		PatRuleID: locality.ExpandID(patRuleIDRaw),
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCPublicGatewayPATRuleRead(ctx, d, m)
}
