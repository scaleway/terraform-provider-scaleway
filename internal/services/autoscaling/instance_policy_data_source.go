package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstancePolicy() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstancePolicy().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "instance_group_id", "zone")

	dsSchema["name"].ConflictsWith = []string{"instance_policy_id"}
	dsSchema["instance_policy_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the instance policy",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceInstancePolicyRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstancePolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := NewAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	policyID, ok := d.GetOk("instance_policy_id")
	if !ok {
		policyName := d.Get("name").(string)

		instanceGroupIDRaw, instanceGroupIDOk := d.GetOk("instance_group_id")
		if !instanceGroupIDOk {
			return diag.Errorf("instance_group_id is required when looking up instance policy by name")
		}

		res, err := api.ListInstancePolicies(&autoscaling.ListInstancePoliciesRequest{
			Zone:            zone,
			InstanceGroupID: locality.ExpandID(instanceGroupIDRaw.(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPolicy, err := datasource.FindExact(
			res.Policies,
			func(p *autoscaling.InstancePolicy) bool { return p.Name == policyName },
			policyName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		policyID = foundPolicy.ID
	}

	zonedID := datasource.NewZonedID(policyID, zone)
	d.SetId(zonedID)

	return ResourceInstancePolicyRead(ctx, d, m)
}
