package autoscaling

import (
	"context"
	_ "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func ResourceInstanceGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := NewAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &autoscaling.CreateInstanceGroupRequest{
		Zone:         zone,
		ProjectID:    d.Get("project_id").(string),
		Name:         types.ExpandOrGenerateString(d.Get("name").(string), "instance-group"),
		Tags:         types.ExpandStrings(d.Get("tags")),
		TemplateID:   locality.ExpandID(d.Get("template_id").(string)),
		Capacity:     expandInstanceCapacity(d.Get("capacity")),
		Loadbalancer: expandInstanceLoadBalancer(d.Get("load_balancer")),
	}

	group, err := api.CreateInstanceGroup(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, group.ID))

	return ResourceInstanceGroupRead(ctx, d, m)
}

func ResourceInstanceGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := api.GetInstanceGroup(&autoscaling.GetInstanceGroupRequest{
		Zone:            zone,
		InstanceGroupID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", group.Name)
	_ = d.Set("template_id", zonal.NewIDString(zone, group.InstanceTemplateID))
	_ = d.Set("tags", group.Tags)
	_ = d.Set("capacity", flattenInstanceCapacity(group.Capacity))
	_ = d.Set("load_balancer", flattenInstanceLoadBalancer(group.Loadbalancer, zone))
	_ = d.Set("created_at", types.FlattenTime(group.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(group.UpdatedAt))
	_ = d.Set("zone", zone)
	_ = d.Set("project_id", group.ProjectID)

	return nil
}

func ResourceInstanceGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &autoscaling.UpdateInstanceGroupRequest{
		Zone:            zone,
		InstanceGroupID: ID,
	}

	hasChanged := false

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("capacity") {
		updateRequest.Capacity = expandUpdateInstanceCapacity(d.Get("capacity"))
		hasChanged = true
	}

	if d.HasChange("load_balancer") {
		updateRequest.Loadbalancer = expandUpdateInstanceLoadBalancer(d.Get("load_balancer"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateInstanceGroup(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstanceGroupRead(ctx, d, m)
}

func ResourceInstanceGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := api.GetInstanceGroup(&autoscaling.GetInstanceGroupRequest{
		Zone:            zone,
		InstanceGroupID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteInstanceGroup(&autoscaling.DeleteInstanceGroupRequest{
		Zone:            zone,
		InstanceGroupID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("delete_servers_on_destroy").(bool) {
		instanceAPI := instanceSDK.NewAPI(meta.ExtractScwClient(m))

		err = instance.DeleteASGServers(ctx, instanceAPI, zone, group.ID, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
