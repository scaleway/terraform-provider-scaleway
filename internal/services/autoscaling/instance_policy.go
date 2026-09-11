package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceInstancePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstancePolicyCreate,
		ReadContext:   ResourceInstancePolicyRead,
		UpdateContext: ResourceInstancePolicyUpdate,
		DeleteContext: ResourceInstancePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion:      0,
		SchemaFunc:         instancePolicySchema,
		DeprecationMessage: deprecationMessage,
	}
}

func instancePolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"instance_group_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "ID of the instance group related to this policy",
			DiffSuppressFunc: dsf.Locality,
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The policy name",
		},
		"action": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "Action to execute when the metric-based condition is met",
			ValidateDiagFunc: verify.ValidateEnum[autoscaling.InstancePolicyAction](),
		},
		"type": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "How to use the number defined in `value` when determining by how many Instances to scale up/down",
			ValidateDiagFunc: verify.ValidateEnum[autoscaling.InstancePolicyType](),
		},
		"value": {
			Type:     schema.TypeInt,
			Required: true,
			Description: "Value representing the magnitude of the scaling action to take for the Instance group. Depending on the `type` parameter, " +
				"this number could represent a total number of Instances in the group, a number of Instances to add, or a percentage to scale the group by",
		},
		"priority": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Priority of this policy compared to all other scaling policies. This determines the processing order. The lower the number, the higher the priority",
		},
		"metric": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cockpit metric to use when determining whether to trigger a scale up/down action",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name or description of the metric policy",
					},
					"operator": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Operator used when comparing the threshold value of the chosen `metric` to the actual sampled and aggregated value",
					},
					"aggregate": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "How the values sampled for the `metric` should be aggregated",
					},
					"managed_metric": {
						Type:     schema.TypeString,
						Optional: true,
						Description: "Managed metric to use for this policy. These are available by default in Cockpit without any configuration or `node_exporter`. " +
							"The chosen metric forms the basis of the condition that will be checked to determine whether a scaling action should be triggered",
					},
					"cockpit_metric_name": {
						Type:     schema.TypeString,
						Optional: true,
						Description: "Custom metric to use for this policy. This must be stored in Scaleway Cockpit. " +
							"The metric forms the basis of the condition that will be checked to determine whether a scaling action should be triggered",
					},
					"sampling_range_min": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Interval of time, in minutes, during which metric is sampled",
					},
					"threshold": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Threshold value to measure the aggregated sampled `metric` value against. Combined with the `operator` field, determines whether a scaling action should be triggered",
					},
				},
			},
		},
		"zone":       zonal.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceInstancePolicyCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_autoscaling_instance_policy is no longer supported",
		Detail:   deprecationMessage,
	}}
}

func ResourceInstancePolicyRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstancePolicyUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstancePolicyDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
