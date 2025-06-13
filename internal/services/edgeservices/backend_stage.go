package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceBackendStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceBackendStageCreate,
		ReadContext:   ResourceBackendStageRead,
		UpdateContext: ResourceBackendStageUpdate,
		DeleteContext: ResourceBackendStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the pipeline",
			},
			"s3_backend_config": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"lb_backend_config"},
				MaxItems:      1,
				Description:   "The Scaleway Object Storage origin bucket (S3) linked to the backend stage",
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
			"lb_backend_config": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"s3_backend_config"},
				Description:   "The Scaleway Load Balancer origin linked to the backend stage",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lb_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The Load Balancer configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "ID of the Load Balancer",
									},
									"frontend_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "ID of the frontend linked to the Load Balancer",
									},
									"is_ssl": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Defines whether the Load Balancer's frontend handles SSL connections",
									},
									"domain_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Fully Qualified Domain Name (in the format subdomain.example.com) to use in HTTP requests sent towards your Load Balancer",
									},
									"zone": zonal.Schema(),
								},
							},
						},
					},
				},
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

func ResourceBackendStageCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := edgeServicesAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &edgeservices.CreateBackendStageRequest{
		PipelineID: d.Get("pipeline_id").(string),
	}

	if s3Config, ok := d.GetOk("s3_backend_config"); ok {
		req.ScalewayS3 = expandS3BackendConfig(s3Config)
	}

	if lbConfig, ok := d.GetOk("lb_backend_config"); ok {
		req.ScalewayLB = expandLBBackendConfig(zone, lbConfig)
	}

	backendStage, err := api.CreateBackendStage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(backendStage.ID)

	return ResourceBackendStageRead(ctx, d, m)
}

func ResourceBackendStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := edgeServicesAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	backendStage, err := api.GetBackendStage(&edgeservices.GetBackendStageRequest{
		BackendStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", backendStage.PipelineID)
	_ = d.Set("created_at", types.FlattenTime(backendStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(backendStage.UpdatedAt))

	if backendStage.ScalewayS3 != nil {
		_ = d.Set("s3_backend_config", flattenS3BackendConfig(backendStage.ScalewayS3))
	}

	if backendStage.ScalewayLB != nil {
		_ = d.Set("lb_backend_config", flattenLBBackendConfig(zone, backendStage.ScalewayLB))
	}

	return nil
}

func ResourceBackendStageUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := edgeServicesAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	updateRequest := &edgeservices.UpdateBackendStageRequest{
		BackendStageID: d.Id(),
	}

	if d.HasChange("s3_backend_config") {
		updateRequest.ScalewayS3 = expandS3BackendConfig(d.Get("s3_backend_config"))
		hasChanged = true
	}

	if d.HasChange("lb_backend_config") {
		updateRequest.ScalewayLB = expandLBBackendConfig(zone, d.Get("lb_backend_config"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateBackendStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceBackendStageRead(ctx, d, m)
}

func ResourceBackendStageDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteBackendStage(&edgeservices.DeleteBackendStageRequest{
		BackendStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
