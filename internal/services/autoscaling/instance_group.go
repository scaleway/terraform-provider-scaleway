package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceInstanceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceGroupCreate,
		ReadContext:   ResourceInstanceGroupRead,
		UpdateContext: ResourceInstanceGroupUpdate,
		DeleteContext: ResourceInstanceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion:      0,
		SchemaFunc:         instanceGroupSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func instanceGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"template_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "ID of the Instance template to attach to the Instance group",
			DiffSuppressFunc: dsf.Locality,
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The Instance group name",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the Instance group",
		},
		"capacity": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The specification of the minimum and maximum replicas for the Instance group, and the cooldown interval between two scaling events",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_replicas": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The maximum count of Instances for the Instance group",
					},
					"min_replicas": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The minimum count of Instances for the Instance group",
					},
					"cooldown_delay": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Time (in seconds) after a scaling action during which requests to carry out a new scaling action will be denied",
					},
				},
			},
		},
		"load_balancer": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The specification of the Load Balancer to link to the Instance group",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:             schema.TypeString,
						Optional:         true,
						Description:      "The ID of the load balancer",
						DiffSuppressFunc: dsf.Locality,
					},
					"backend_ids": {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: dsf.Locality,
						},
						Optional:    true,
						Description: "The Load Balancer backend IDs",
					},
					"private_network_id": {
						Type:             schema.TypeString,
						Optional:         true,
						Description:      "The ID of the Private Network attached to the Load Balancer",
						DiffSuppressFunc: dsf.Locality,
					},
				},
			},
		},
		"delete_servers_on_destroy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to delete all instances in this group when the group is destroyed. Set to `true` to tear them down, `false` (the default) leaves them running",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the Instance group",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the Instance group",
		},
		"zone":       zonal.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceInstanceGroupCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_autoscaling_instance_group is no longer supported",
		Detail:   deprecationMessage,
	}}
}

func ResourceInstanceGroupRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstanceGroupUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstanceGroupDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
