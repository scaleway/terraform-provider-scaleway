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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
		SchemaVersion: 0,
		SchemaFunc:    instanceTemplateSchema,
		Identity:      identity.DefaultZonal(),
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

func ResourceInstanceTemplateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := NewAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &autoscaling.CreateInstanceTemplateRequest{
		Zone:              zone,
		CommercialType:    d.Get("commercial_type").(string),
		ImageID:           types.ExpandStringPtr(locality.ExpandID(d.Get("image_id"))),
		Tags:              types.ExpandStrings(d.Get("tags")),
		SecurityGroupID:   types.ExpandStringPtr(locality.ExpandID(d.Get("security_group_id"))),
		PlacementGroupID:  types.ExpandStringPtr(locality.ExpandID(d.Get("placement_group_id"))),
		PublicIPsV4Count:  types.ExpandUint32Ptr(d.Get("public_ips_v4_count")),
		PublicIPsV6Count:  types.ExpandUint32Ptr(d.Get("public_ips_v6_count")),
		ProjectID:         d.Get("project_id").(string),
		Name:              types.ExpandOrGenerateString(d.Get("name").(string), "template"),
		PrivateNetworkIDs: locality.ExpandIDs(d.Get("private_network_ids")),
	}

	if ci, ok := d.GetOk("cloud_init"); ok {
		rawCI := []byte(ci.(string))
		req.CloudInit = &rawCI
	}

	volumesList := expandVolumes(d.Get("volumes").([]any))

	req.Volumes = volumesList

	template, err := api.CreateInstanceTemplate(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, template.Zone, template.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceTemplateRead(ctx, d, m)
}

func ResourceInstanceTemplateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	template, err := api.GetInstanceTemplate(&autoscaling.GetInstanceTemplateRequest{
		Zone:       zone,
		TemplateID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	pnRegion, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", template.Name)
	_ = d.Set("commercial_type", template.CommercialType)
	_ = d.Set("tags", template.Tags)
	_ = d.Set("public_ips_v4_count", types.FlattenUint32Ptr(template.PublicIPsV4Count))
	_ = d.Set("public_ips_v6_count", types.FlattenUint32Ptr(template.PublicIPsV6Count))
	_ = d.Set("private_network_ids", regional.NewIDStrings(pnRegion, template.PrivateNetworkIDs))
	_ = d.Set("status", template.Status.String())
	_ = d.Set("created_at", types.FlattenTime(template.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(template.UpdatedAt))
	_ = d.Set("zone", zone)
	_ = d.Set("project_id", template.ProjectID)
	_ = d.Set("volumes", flattenVolumes(zone, template.Volumes))

	if template.SecurityGroupID != nil {
		_ = d.Set("security_group_id", zonal.NewIDString(zone, types.FlattenStringPtr(template.SecurityGroupID).(string)))
	}

	if template.PlacementGroupID != nil {
		_ = d.Set("placement_group_id", zonal.NewIDString(zone, types.FlattenStringPtr(template.PlacementGroupID).(string)))
	}

	if template.ImageID != nil {
		_ = d.Set("image_id", zonal.NewIDString(zone, types.FlattenStringPtr(template.ImageID).(string)))
	}

	if template.CloudInit != nil {
		_ = d.Set("cloud_init", string(*template.CloudInit))
	}

	return nil
}

func ResourceInstanceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &autoscaling.UpdateInstanceTemplateRequest{
		Zone:       zone,
		TemplateID: ID,
	}

	hasChanged := false

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("commercial_type") {
		updateRequest.CommercialType = types.ExpandUpdatedStringPtr(d.Get("commercial_type"))
		hasChanged = true
	}

	if d.HasChange("image_id") {
		updateRequest.ImageID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("image_id")))
		hasChanged = true
	}

	if d.HasChange("security_group_id") {
		updateRequest.SecurityGroupID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("security_group_id")))
		hasChanged = true
	}

	if d.HasChange("placement_group_id") {
		updateRequest.PlacementGroupID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("placement_group_id")))
		hasChanged = true
	}

	if d.HasChange("public_ips_v4_count") {
		updateRequest.PublicIPsV4Count = types.ExpandUint32Ptr(d.Get("public_ips_v4_count"))
		hasChanged = true
	}

	if d.HasChange("public_ips_v6_count") {
		updateRequest.PublicIPsV6Count = types.ExpandUint32Ptr(d.Get("public_ips_v6_count"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("private_network_ids") {
		updateRequest.PrivateNetworkIDs = types.ExpandUpdatedStringsPtr(locality.ExpandIDs(d.Get("private_network_ids")))
		hasChanged = true
	}

	if d.HasChange("cloud_init") {
		rawCI := []byte(d.Get("cloud_init").(string))
		updateRequest.CloudInit = &rawCI
		hasChanged = true
	}

	if d.HasChange("volumes") {
		updateRequest.Volumes = expandVolumes(d.Get("volumes").([]any))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateInstanceTemplate(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstanceTemplateRead(ctx, d, m)
}

func ResourceInstanceTemplateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteInstanceTemplate(&autoscaling.DeleteInstanceTemplateRequest{
		Zone:       zone,
		TemplateID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
