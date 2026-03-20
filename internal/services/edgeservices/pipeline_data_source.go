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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/pipeline_data_source.md
var pipelineDataSourceDescription string

func DataSourcePipeline() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePipeline().SchemaFunc())

	filterFields := []string{"name", "project_id"}

	dsSchema["pipeline_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the pipeline",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "project_id")

	dsSchema["name"].ConflictsWith = []string{"pipeline_id"}
	dsSchema["project_id"].ConflictsWith = []string{"pipeline_id"}

	return &schema.Resource{
		ReadContext: DataSourcePipelineRead,
		Description: pipelineDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourcePipelineRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	pipelineID, idExists := d.GetOk("pipeline_id")
	if idExists {
		return dataSourcePipelineReadByID(ctx, d, m, pipelineID.(string))
	}

	return dataSourcePipelineReadByFilters(ctx, d, m)
}

func dataSourcePipelineReadByID(ctx context.Context, d *schema.ResourceData, m any, pipelineID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(pipelineID)

	pipeline, err := api.GetPipeline(&edgeservices.GetPipelineRequest{
		PipelineID: pipelineID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setPipelineState(d, pipeline)
}

func dataSourcePipelineReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	req := &edgeservices.ListPipelinesRequest{
		Name:      types.ExpandStringPtr(d.Get("name")),
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
	}

	res, err := api.ListPipelines(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Pipelines) == 0 {
		return diag.FromErr(errors.New("no pipeline found matching the specified filters"))
	}

	if len(res.Pipelines) > 1 {
		return diag.FromErr(fmt.Errorf("multiple pipelines (%d) found, please refine your filters to match exactly one", len(res.Pipelines)))
	}

	pipeline := res.Pipelines[0]
	d.SetId(pipeline.ID)

	return setPipelineState(d, pipeline)
}
