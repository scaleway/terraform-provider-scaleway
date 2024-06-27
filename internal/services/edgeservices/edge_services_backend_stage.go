package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/edge_services/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceEdgeServicesBackendStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceEdgeServicesBackendStageCreate,
		ReadContext:   ResourceEdgeServicesBackendStageRead,
		UpdateContext: ResourceEdgeServicesBackendStageUpdate,
		DeleteContext: ResourceEdgeServicesBackendStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"s3_backend_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Scaleway Object Storage origin bucket (S3) linked to the backend stage",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the Bucket",
						},
						"bucket_region": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The region of the Bucket",
						},
						"is_website": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Defines whether the bucket website feature is enabled.",
						},
					},
				},
			},
			"pipeline_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pipeline ID the backend stage belongs to",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the backend stage",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the backend stage",
			},
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceEdgeServicesBackendStageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	backendStage, err := api.CreateBackendStage(&edge_services.CreateBackendStageRequest{
		ProjectID:  d.Get("project_id").(string),
		ScalewayS3: expandEdgeServicesScalewayS3BackendConfig(d.Get("s3_backend_config")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(backendStage.ID)

	return ResourceEdgeServicesBackendStageRead(ctx, d, m)
}

func ResourceEdgeServicesBackendStageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	backendStage, err := api.GetBackendStage(&edge_services.GetBackendStageRequest{
		BackendStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", types.FlattenStringPtr(backendStage.PipelineID))
	_ = d.Set("created_at", types.FlattenTime(backendStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(backendStage.UpdatedAt))
	_ = d.Set("project_id", backendStage.ProjectID)
	_ = d.Set("s3_backend_config", flattenEdgeServicesScalewayS3BackendConfig(backendStage.ScalewayS3))

	return nil
}

func ResourceEdgeServicesBackendStageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edge_services.UpdateBackendStageRequest{
		BackendStageID: d.Id(),
	}

	if d.HasChange("s3_backend_config") {
		updateRequest.ScalewayS3 = expandEdgeServicesScalewayS3BackendConfig(d.Get("s3_backend_config"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateBackendStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceEdgeServicesBackendStageRead(ctx, d, m)
}

func ResourceEdgeServicesBackendStageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteBackendStage(&edge_services.DeleteBackendStageRequest{
		BackendStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is403(err) {
		return diag.FromErr(err)
	}

	return nil
}
