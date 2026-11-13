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

//go:embed descriptions/cache_stage_data_source.md
var cacheStageDataSourceDescription string

func DataSourceCacheStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCacheStage().SchemaFunc())

	filterFields := []string{"pipeline_id"}

	dsSchema["cache_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the cache stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")

	dsSchema["pipeline_id"].ConflictsWith = []string{"cache_stage_id"}

	return &schema.Resource{
		ReadContext: DataSourceCacheStageRead,
		Description: cacheStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceCacheStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("cache_stage_id")
	if idExists {
		return dataSourceCacheStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceCacheStageReadByFilters(ctx, d, m)
}

func dataSourceCacheStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(stageID)

	cacheStage, err := api.GetCacheStage(&edgeservices.GetCacheStageRequest{
		CacheStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setCacheStageState(d, cacheStage)
}

func dataSourceCacheStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when cache_stage_id is not specified"))
	}

	res, err := api.ListCacheStages(&edgeservices.ListCacheStagesRequest{
		PipelineID: pipelineID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no cache stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple cache stages (%d) found, please refine your filters or use cache_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)

	return setCacheStageState(d, stage)
}
