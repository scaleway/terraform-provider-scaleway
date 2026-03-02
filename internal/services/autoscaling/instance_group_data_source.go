package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstanceGroup() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstanceGroup().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"instance_group_id"}
	dsSchema["instance_group_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the instance group",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceInstanceGroupRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstanceGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := NewAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceGroupID, ok := d.GetOk("instance_group_id")
	if !ok {
		instanceGroupName := d.Get("name").(string)

		res, err := api.ListInstanceGroups(&autoscaling.ListInstanceGroupsRequest{
			Zone: zone,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundGroup, err := datasource.FindExact(
			res.InstanceGroups,
			func(g *autoscaling.InstanceGroup) bool {
				return g.Name == instanceGroupName
			},
			instanceGroupName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		instanceGroupID = foundGroup.ID
	}

	zonedID := datasource.NewZonedID(instanceGroupID, zone)
	d.SetId(zonedID)

	return ResourceInstanceGroupRead(ctx, d, m)
}
