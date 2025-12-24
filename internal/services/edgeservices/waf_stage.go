package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceWAFStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceWAFStageCreate,
		ReadContext:   ResourceWAFStageRead,
		UpdateContext: ResourceWAFStageUpdate,
		DeleteContext: ResourceWAFStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    wafStageSchema,
		Identity: identity.WrapSchemaMap(map[string]*schema.Schema{
			"waf_stage_id": {
				Type:              schema.TypeString,
				Description:       "The ID of the WAF Stage (UUID format)",
				RequiredForImport: true,
			},
		}),
	}
}

func wafStageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"pipeline_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of the pipeline",
		},
		"backend_stage_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The ID of the backend stage to forward requests to after the WAF stage",
		},
		"paranoia_level": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The sensitivity level (`1`,`2`,`3`,`4`) to use when classifying requests as malicious. With a high level, requests are more likely to be classed as malicious, and false positives are expected. With a lower level, requests are more likely to be classed as benign",
		},
		"mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Mode defining WAF behavior (`disable`/`log_only`/`enable`)",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the WAF stage",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the WAF stage",
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceWAFStageCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	wafStage, err := api.CreateWafStage(&edgeservices.CreateWafStageRequest{
		PipelineID:     d.Get("pipeline_id").(string),
		BackendStageID: types.ExpandStringPtr(d.Get("backend_stage_id").(string)),
		ParanoiaLevel:  uint32(d.Get("paranoia_level").(int)),
		Mode:           edgeservices.WafStageMode(d.Get("mode").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(wafStage.ID)

	return ResourceWAFStageRead(ctx, d, m)
}

func ResourceWAFStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	wafStage, err := api.GetWafStage(&edgeservices.GetWafStageRequest{
		WafStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", wafStage.PipelineID)
	_ = d.Set("backend_stage_id", types.FlattenStringPtr(wafStage.BackendStageID))
	_ = d.Set("paranoia_level", int(wafStage.ParanoiaLevel))
	_ = d.Set("mode", wafStage.Mode.String())
	_ = d.Set("created_at", types.FlattenTime(wafStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(wafStage.UpdatedAt))

	return nil
}

func ResourceWAFStageUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edgeservices.UpdateWafStageRequest{
		WafStageID: d.Id(),
	}

	if d.HasChange("mode") {
		updateRequest.Mode = edgeservices.WafStageMode(d.Get("mode").(string))
		hasChanged = true
	}

	if d.HasChange("paranoia_level") {
		updateRequest.ParanoiaLevel = types.ExpandUint32Ptr(d.Get("paranoia_level"))
		hasChanged = true
	}

	if d.HasChange("backend_stage_id") {
		updateRequest.BackendStageID = types.ExpandStringPtr(d.Get("backend_stage_id").(string))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateWafStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceWAFStageRead(ctx, d, m)
}

func ResourceWAFStageDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteWafStage(&edgeservices.DeleteWafStageRequest{
		WafStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
