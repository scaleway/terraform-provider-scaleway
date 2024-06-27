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

func ResourceEdgeServicesCacheStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceEdgeServicesCacheStageCreate,
		ReadContext:   ResourceEdgeServicesCacheStageRead,
		UpdateContext: ResourceEdgeServicesCacheStageUpdate,
		DeleteContext: ResourceEdgeServicesCacheStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
			"pipeline_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pipeline description",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pipeline description",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pipeline description",
			},
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceEdgeServicesCacheStageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	cacheStage, err := api.CreateCacheStage(&edge_services.CreateCacheStageRequest{
		ProjectID:      d.Get("project_id").(string),
		BackendStageID: types.ExpandStringPtr(d.Get("backend_stage_id").(string)),
		FallbackTTL:    &scw.Duration{Seconds: 3600},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cacheStage.ID)

	return ResourceEdgeServicesCacheStageRead(ctx, d, m)
}

func ResourceEdgeServicesCacheStageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	cacheStage, err := api.GetCacheStage(&edge_services.GetCacheStageRequest{
		CacheStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", types.FlattenStringPtr(cacheStage.PipelineID))
	_ = d.Set("project_id", cacheStage.ProjectID)
	_ = d.Set("created_at", types.FlattenTime(cacheStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(cacheStage.UpdatedAt))
	_ = d.Set("backend_stage_id", types.FlattenStringPtr(cacheStage.BackendStageID))
	_ = d.Set("fallback_ttl", cacheStage.FallbackTTL.Seconds)

	/*	prs, err := api.ListPurgeRequests(&edge_services.ListPurgeRequestsRequest{
			ProjectID:  scw.StringPtr(cacheStage.ProjectID),
			PipelineID: cacheStage.PipelineID,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("purge", flattenEdgeServicesPurge(prs.PurgeRequests))
	*/
	return nil
}

func ResourceEdgeServicesCacheStageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edge_services.UpdateCacheStageRequest{
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
		for _, pr := range expandEdgeServicesPurge(d.Get("purge_requests")) {
			res, err := api.CreatePurgeRequest(&edge_services.CreatePurgeRequestRequest{
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

	return ResourceEdgeServicesCacheStageRead(ctx, d, m)
}

func ResourceEdgeServicesCacheStageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteCacheStage(&edge_services.DeleteCacheStageRequest{
		CacheStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is403(err) {
		return diag.FromErr(err)
	}

	return nil
}
