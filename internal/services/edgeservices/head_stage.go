package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

func ResourceHeadStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceHeadStageCreate,
		ReadContext:   ResourceHeadStageRead,
		UpdateContext: ResourceHeadStageUpdate,
		DeleteContext: ResourceHeadStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the pipeline ID",
			},
			"head_stage_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the head stage of the pipeline",
			},
		},
	}
}

func ResourceHeadStageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	dnsStage, err := api.SetHeadStage(&edgeservices.SetHeadStageRequest{
		PipelineID: d.Get("pipeline_id").(string),
		AddNewHeadStage: &edgeservices.SetHeadStageRequestAddNewHeadStage{
			NewStageID: d.Get("head_stage_id").(string),
		},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*dnsStage.HeadStage.DNSStageID)

	return ResourceHeadStageRead(ctx, d, m)
}

func ResourceHeadStageRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func ResourceHeadStageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	if d.HasChange("head_stage_id") {
		before, after := d.GetChange("head_stage_id")

		_, err := api.SetHeadStage(&edgeservices.SetHeadStageRequest{
			PipelineID: d.Get("pipeline_id").(string),
			SwapHeadStage: &edgeservices.SetHeadStageRequestSwapHeadStage{
				CurrentStageID: before.(string),
				NewStageID:     after.(string),
			},
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceHeadStageRead(ctx, d, m)
}

func ResourceHeadStageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	_, err := api.SetHeadStage(&edgeservices.SetHeadStageRequest{
		PipelineID: d.Get("pipeline_id").(string),
		RemoveHeadStage: &edgeservices.SetHeadStageRequestRemoveHeadStage{
			RemoveStageID: d.Get("head_stage_id").(string),
		},
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
