package edgeservices

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/waf_stage_data_source.md
var wafStageDataSourceDescription string

func DataSourceWAFStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceWAFStage().SchemaFunc())

	filterFields := []string{"pipeline_id"}

	dsSchema["waf_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the WAF stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")

	dsSchema["pipeline_id"].ConflictsWith = []string{"waf_stage_id"}

	return &schema.Resource{
		ReadContext: DataSourceWAFStageRead,
		Description: wafStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceWAFStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("waf_stage_id")
	if idExists {
		return dataSourceWAFStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceWAFStageReadByFilters(ctx, d, m)
}

func dataSourceWAFStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(stageID)

	wafStage, err := api.GetWafStage(&edgeservices.GetWafStageRequest{
		WafStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setWAFStageState(d, wafStage)
}

func dataSourceWAFStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when waf_stage_id is not specified"))
	}

	res, err := api.ListWafStages(&edgeservices.ListWafStagesRequest{
		PipelineID: pipelineID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no WAF stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple WAF stages (%d) found, please refine your filters or use waf_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)

	return setWAFStageState(d, stage)
}
