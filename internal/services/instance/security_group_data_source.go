package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceSecurityGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSecurityGroup().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"security_group_id"}
	dsSchema["security_group_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the security group",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceInstanceSecurityGroupRead,

		Schema: dsSchema,
	}
}

func DataSourceInstanceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	securityGroupID, ok := d.GetOk("security_group_id")
	if !ok {
		sgName := d.Get("name").(string)
		res, err := instanceAPI.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(sgName),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundSG, err := datasource.FindExact(
			res.SecurityGroups,
			func(s *instance.SecurityGroup) bool { return s.Name == sgName },
			sgName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		securityGroupID = foundSG.ID
	}

	zonedID := datasource.NewZonedID(securityGroupID, zone)
	d.SetId(zonedID)
	_ = d.Set("security_group_id", zonedID)
	return ResourceInstanceSecurityGroupRead(ctx, d, m)
}
