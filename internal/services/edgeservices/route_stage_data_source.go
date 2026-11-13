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

//go:embed descriptions/route_stage_data_source.md
var routeStageDataSourceDescription string

func DataSourceRouteStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRouteStage().SchemaFunc())

	filterFields := []string{"pipeline_id"}

	dsSchema["route_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the route stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")

	dsSchema["pipeline_id"].ConflictsWith = []string{"route_stage_id"}

	return &schema.Resource{
		ReadContext: DataSourceRouteStageRead,
		Description: routeStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceRouteStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("route_stage_id")
	if idExists {
		return dataSourceRouteStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceRouteStageReadByFilters(ctx, d, m)
}

func dataSourceRouteStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(stageID)

	routeStage, err := api.GetRouteStage(&edgeservices.GetRouteStageRequest{
		RouteStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	routeRules, err := api.ListRouteRules(&edgeservices.ListRouteRulesRequest{
		RouteStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setRouteStageState(d, routeStage, routeRules.RouteRules)
}

func dataSourceRouteStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when route_stage_id is not specified"))
	}

	res, err := api.ListRouteStages(&edgeservices.ListRouteStagesRequest{
		PipelineID: pipelineID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no route stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple route stages (%d) found, please refine your filters or use route_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)

	routeRules, err := api.ListRouteRules(&edgeservices.ListRouteRulesRequest{
		RouteStageID: stage.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setRouteStageState(d, stage, routeRules.RouteRules)
}
