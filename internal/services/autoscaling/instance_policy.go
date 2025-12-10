package autoscaling

import (
	"context"
	_ "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
		SchemaVersion: 0,
		SchemaFunc:    instancePolicySchema,
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

func ResourceInstancePolicyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := NewAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &autoscaling.CreateInstancePolicyRequest{
		Zone:            zone,
		InstanceGroupID: locality.ExpandID(d.Get("instance_group_id").(string)),
		Name:            types.ExpandOrGenerateString(d.Get("name").(string), "instance-policy"),
		Action:          autoscaling.InstancePolicyAction(d.Get("action").(string)),
		Type:            autoscaling.InstancePolicyType(d.Get("type").(string)),
		Value:           uint32(d.Get("value").(int)),
		Priority:        uint32(d.Get("priority").(int)),
		Metric:          expandPolicyMetric(d.Get("metric").([]any)),
	}

	policy, err := api.CreateInstancePolicy(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, policy.ID))

	return ResourceInstancePolicyRead(ctx, d, m)
}

func ResourceInstancePolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := api.GetInstancePolicy(&autoscaling.GetInstancePolicyRequest{
		Zone:     zone,
		PolicyID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", policy.Name)
	_ = d.Set("action", policy.Action.String())
	_ = d.Set("type", policy.Type.String())
	_ = d.Set("value", int(policy.Value))
	_ = d.Set("priority", int(policy.Priority))
	_ = d.Set("metric", flattenPolicyMetric(policy.Metric))
	_ = d.Set("instance_group_id", zonal.NewIDString(zone, policy.InstanceGroupID))
	_ = d.Set("zone", zone)

	return nil
}

func ResourceInstancePolicyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &autoscaling.UpdateInstancePolicyRequest{
		Zone:     zone,
		PolicyID: ID,
		Action:   autoscaling.InstancePolicyAction(d.Get("action").(string)),
		Type:     autoscaling.InstancePolicyType(d.Get("type").(string)),
	}

	hasChanged := false

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("metric") {
		updateRequest.Metric = expandUpdatePolicyMetric(d.Get("metric"))
		hasChanged = true
	}

	if d.HasChange("value") {
		updateRequest.Value = types.ExpandUint32Ptr(d.Get("value"))
		hasChanged = true
	}

	if d.HasChange("priority") {
		updateRequest.Priority = types.ExpandUint32Ptr(d.Get("priority"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateInstancePolicy(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstancePolicyRead(ctx, d, m)
}

func ResourceInstancePolicyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteInstancePolicy(&autoscaling.DeleteInstancePolicyRequest{
		Zone:     zone,
		PolicyID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
