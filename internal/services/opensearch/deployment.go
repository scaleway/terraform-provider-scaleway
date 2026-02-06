package opensearch

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeploymentCreate,
		ReadContext:   resourceDeploymentRead,
		UpdateContext: resourceDeploymentUpdate,
		DeleteContext: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		SchemaFunc: deploymentSchema,
	}
}

func deploymentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Name of the OpenSearch deployment",
		},
		"tags": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "List of tags to apply",
		},
		"version": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "OpenSearch version to use",
		},
		"node_amount": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of nodes",
		},
		"node_type": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Type of node",
		},
		"user_name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Username for the deployment",
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Optional:    true,
			ForceNew:    true,
			Description: "Password for the deployment user",
		},
		"volume": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Volume configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
						Description: "Volume type (sbs_5k, sbs_15k)",
					},
					"size_bytes": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Volume size in bytes",
					},
				},
			},
		},
		"endpoints": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of endpoints",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Endpoint ID",
					},
					"services": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "List of services",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service name",
								},
								"port": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Service port",
								},
								"url": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service URL",
								},
							},
						},
					},
					"public": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the endpoint is public",
					},
					"private_network_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Private network ID if applicable",
					},
				},
			},
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the deployment",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of deployment creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of deployment last update (RFC 3339 format)",
		},
	}
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &searchdbapi.CreateDeploymentRequest{
		Region:     region,
		ProjectID:  d.Get("project_id").(string),
		Name:       types.ExpandOrGenerateString(d.Get("name"), "opensearch"),
		Version:    d.Get("version").(string),
		NodeAmount: uint32(d.Get("node_amount").(int)),
		NodeType:   d.Get("node_type").(string),
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(v)
	}

	if v, ok := d.GetOk("user_name"); ok {
		req.UserName = types.ExpandStringPtr(v)
	}

	if v, ok := d.GetOk("password"); ok {
		req.Password = types.ExpandStringPtr(v)
	}

	if v, ok := d.GetOk("volume"); ok {
		volumeList := v.([]any)
		if len(volumeList) > 0 {
			volumeMap := volumeList[0].(map[string]any)
			req.Volume = &searchdbapi.Volume{
				Type:      searchdbapi.VolumeType(volumeMap["type"].(string)),
				SizeBytes: scw.Size(uint64(volumeMap["size_bytes"].(int))),
			}
		}
	}

	// Create public endpoint by default
	req.Endpoints = []*searchdbapi.EndpointSpec{
		{
			Public: &searchdbapi.EndpointSpecPublicDetails{},
		},
	}

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err = waitForDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, deployment.ID))

	return resourceDeploymentRead(ctx, d, meta)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := api.GetDeployment(&searchdbapi.GetDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("region", string(deployment.Region))
	_ = d.Set("project_id", deployment.ProjectID)
	_ = d.Set("name", deployment.Name)
	_ = d.Set("tags", types.FlattenSliceString(deployment.Tags))
	_ = d.Set("version", deployment.Version)
	_ = d.Set("node_amount", int(deployment.NodeAmount))
	_ = d.Set("node_type", deployment.NodeType)
	_ = d.Set("status", string(deployment.Status))

	if deployment.CreatedAt != nil {
		_ = d.Set("created_at", deployment.CreatedAt.Format(time.RFC3339))
	}

	if deployment.UpdatedAt != nil {
		_ = d.Set("updated_at", deployment.UpdatedAt.Format(time.RFC3339))
	}

	if deployment.Volume != nil {
		_ = d.Set("volume", []map[string]any{
			{
				"type":       string(deployment.Volume.Type),
				"size_bytes": int(deployment.Volume.SizeBytes),
			},
		})
	}

	_ = d.Set("endpoints", flattenEndpoints(deployment.Endpoints))

	return nil
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	// Handle standard updates (name, tags)
	if d.HasChanges("name", "tags") {
		req := &searchdbapi.UpdateDeploymentRequest{
			Region:       region,
			DeploymentID: id,
		}
		changed := false

		if d.HasChange("name") {
			req.Name = types.ExpandStringPtr(d.Get("name"))
			changed = true
		}

		if d.HasChange("tags") {
			req.Tags = scw.StringsPtr(types.ExpandStrings(d.Get("tags")))
			changed = true
		}

		if changed {
			_, err := api.UpdateDeployment(req, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Handle upgrade operations (node_amount, volume size)
	if d.HasChange("node_amount") {
		upgradeReq := &searchdbapi.UpgradeDeploymentRequest{
			Region:       region,
			DeploymentID: id,
			NodeAmount:   scw.Uint32Ptr(uint32(d.Get("node_amount").(int))),
		}

		_, err := api.UpgradeDeployment(upgradeReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("volume") {
		oldVolume, newVolume := d.GetChange("volume")
		oldList := oldVolume.([]any)
		newList := newVolume.([]any)

		if len(oldList) > 0 && len(newList) > 0 {
			oldVolumeMap := oldList[0].(map[string]any)
			newVolumeMap := newList[0].(map[string]any)

			oldSize := uint64(oldVolumeMap["size_bytes"].(int))
			newSize := uint64(newVolumeMap["size_bytes"].(int))

			if oldSize != newSize {
				upgradeReq := &searchdbapi.UpgradeDeploymentRequest{
					Region:          region,
					DeploymentID:    id,
					VolumeSizeBytes: scw.Uint64Ptr(newSize),
				}

				_, err := api.UpgradeDeployment(upgradeReq, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}

				_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	readDiags := resourceDeploymentRead(ctx, d, meta)

	return append(diags, readDiags...)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = api.DeleteDeployment(&searchdbapi.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

