package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceCacheStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCacheStageCreate,
		ReadContext:   ResourceCacheStageRead,
		UpdateContext: ResourceCacheStageUpdate,
		DeleteContext: ResourceCacheStageDelete,
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
			"backend_stage_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The backend stage ID the cache stage will be linked to",
			},
			"fallback_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
				Description: "The Time To Live (TTL) in seconds. Defines how long content is cached",
			},
			"purge_requests": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The pipeline ID in which the purge request will be created",
						},
						"assets": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The list of asserts to purge",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"all": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Defines whether to purge all content",
						},
					},
				},
			},
			"refresh_cache": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Trigger a refresh of the cache by changing this field's value",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the cache stage",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the cache stage",
			},
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceCacheStageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	cacheStage, err := api.CreateCacheStage(&edgeservices.CreateCacheStageRequest{
		PipelineID:     d.Get("pipeline_id").(string),
		BackendStageID: types.ExpandStringPtr(d.Get("backend_stage_id").(string)),
		FallbackTTL:    &scw.Duration{Seconds: int64(d.Get("fallback_ttl").(int))},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cacheStage.ID)

	return ResourceCacheStageRead(ctx, d, m)
}

func ResourceCacheStageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	cacheStage, err := api.GetCacheStage(&edgeservices.GetCacheStageRequest{
		CacheStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", cacheStage.PipelineID)
	_ = d.Set("created_at", types.FlattenTime(cacheStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(cacheStage.UpdatedAt))
	_ = d.Set("backend_stage_id", types.FlattenStringPtr(cacheStage.BackendStageID))
	_ = d.Set("fallback_ttl", cacheStage.FallbackTTL.Seconds)

	return nil
}

func ResourceCacheStageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edgeservices.UpdateCacheStageRequest{
		CacheStageID: d.Id(),
	}

	if d.HasChange("backend_stage_id") {
		updateRequest.BackendStageID = types.ExpandUpdatedStringPtr(d.Get("backend_stage_id"))
		hasChanged = true
	}

	if d.HasChange("fallback_ttl") {
		updateRequest.FallbackTTL = &scw.Duration{Seconds: int64(d.Get("fallback_ttl").(int))}
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateCacheStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("purge_requests", "refresh_cache") {
		for _, pr := range expandPurge(d.Get("purge_requests")) {
			res, err := api.CreatePurgeRequest(&edgeservices.CreatePurgeRequestRequest{
				PipelineID: pr.PipelineID,
				Assets:     pr.Assets,
				All:        pr.All,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForPurge(ctx, api, res.ID, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return ResourceCacheStageRead(ctx, d, m)
}

func ResourceCacheStageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteCacheStage(&edgeservices.DeleteCacheStageRequest{
		CacheStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
