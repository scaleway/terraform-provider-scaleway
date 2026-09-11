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

func ResourceInstanceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceTemplateCreate,
		ReadContext:   ResourceInstanceTemplateRead,
		UpdateContext: ResourceInstanceTemplateUpdate,
		DeleteContext: ResourceInstanceTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion:      0,
		SchemaFunc:         instanceTemplateSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func instanceTemplateSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The Instance template name",
		},
		"commercial_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of Instance commercial type",
		},
		"image_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Instance image ID. Can be an ID of a marketplace or personal image. This image must be compatible with `volume` and `commercial_type` template",
			DiffSuppressFunc: dsf.Locality,
		},
		"security_group_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Instance security group ID",
			DiffSuppressFunc: dsf.Locality,
		},
		"placement_group_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Instance placement group ID. This is optional, but it is highly recommended to set a preference for Instance location within Availability Zone",
			DiffSuppressFunc: dsf.Locality,
		},
		"volumes": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The IPv4 subnet associated with the private network",
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The name of the volume",
					},
					"perf_iops": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The maximum IO/s expected, according to the different options available in stock (`5000 | 15000`)",
					},
					"from_empty": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Volume instance template from empty",
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"size": {
									Type:        schema.TypeInt,
									Required:    true,
									Description: "Size in GB of the new empty volume",
								},
							},
						},
					},
					"from_snapshot": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Volume instance template from snapshot",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"snapshot_id": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "ID of the snapshot to clone",
								},
								"size": {
									Type:        schema.TypeInt,
									Optional:    true,
									Description: "Override size (in GB) of the cloned volume",
								},
							},
						},
					},
					"tags": {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional:    true,
						Description: "List of tags assigned to the volume",
					},
					"boot": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Force the Instance to boot on this volume",
					},
					"volume_type": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "Type of the volume",
						ValidateDiagFunc: verify.ValidateEnum[autoscaling.VolumeInstanceTemplateVolumeType](),
					},
				},
			},
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the Instance template",
		},
		"private_network_ids": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				DiffSuppressFunc: dsf.Locality,
			},
			Optional:    true,
			Description: "Private Network IDs to attach to the new Instance",
		},
		"public_ips_v4_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Number of flexible IPv4 addresses to attach to the new Instance",
		},
		"public_ips_v6_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Number of flexible IPv6 addresses to attach to the new Instance",
		},
		"cloud_init": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Cloud-config to apply to each instance (will be passed in Base64 format)",
		},
		"status": {
			Type:        schema.TypeString,
			Description: "The Instance template status",
			Computed:    true,
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the Instance template",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the Instance template",
		},
		"zone":       zonal.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceInstanceTemplateCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_autoscaling_instance_template is no longer supported",
		Detail:   deprecationMessage,
	}}
}

func ResourceInstanceTemplateRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstanceTemplateUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func ResourceInstanceTemplateDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
