package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayVPCPublicGatewayPATRule() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPublicGatewayPATRule().Schema)

	dsSchema["pat_rule_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The ID of the public gateway PAT rule",
		ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "zone")

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPublicGatewayPATRuleRead,
	}
}

func dataSourceScalewayVPCPublicGatewayPATRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	patRuleIDRaw := d.Get("pat_rule_id")

	zonedID := datasourceNewZonedID(patRuleIDRaw, zone)
	d.SetId(zonedID)
	_ = d.Set("pat_rule_id", zonedID)

	// check if pat rule exist
	_, err = vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: locality.ExpandID(patRuleIDRaw),
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPublicGatewayPATRuleRead(ctx, d, m)
}
