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

//go:embed descriptions/backend_stage_data_source.md
var backendStageDataSourceDescription string

func DataSourceBackendStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceBackendStage().SchemaFunc())

	filterFields := []string{"pipeline_id", "bucket_name", "bucket_region", "lb_id"}

	dsSchema["backend_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the backend stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")
	dsSchema["pipeline_id"].ConflictsWith = []string{"backend_stage_id"}
	dsSchema["bucket_name"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Filter by S3 bucket name",
		ConflictsWith: []string{"backend_stage_id"},
	}
	dsSchema["bucket_region"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Filter by S3 bucket region",
		ConflictsWith: []string{"backend_stage_id"},
	}
	dsSchema["lb_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Filter by Load Balancer ID",
		ConflictsWith: []string{"backend_stage_id"},
	}

	return &schema.Resource{
		ReadContext: DataSourceBackendStageRead,
		Description: backendStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceBackendStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("backend_stage_id")
	if idExists {
		return dataSourceBackendStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceBackendStageReadByFilters(ctx, d, m)
}

func dataSourceBackendStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api, zone, err := edgeServicesAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(stageID)

	backendStage, err := api.GetBackendStage(&edgeservices.GetBackendStageRequest{
		BackendStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setBackendStageState(d, backendStage, zone)
}

func dataSourceBackendStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := edgeServicesAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when backend_stage_id is not specified"))
	}

	req := &edgeservices.ListBackendStagesRequest{
		PipelineID: pipelineID.(string),
	}

	if bucketName, ok := d.GetOk("bucket_name"); ok {
		req.BucketName = types.ExpandStringPtr(bucketName)
	}

	if bucketRegion, ok := d.GetOk("bucket_region"); ok {
		req.BucketRegion = types.ExpandStringPtr(bucketRegion)
	}

	if lbID, ok := d.GetOk("lb_id"); ok {
		req.LBID = types.ExpandStringPtr(lbID)
	}

	res, err := api.ListBackendStages(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no backend stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple backend stages (%d) found, please refine your filters or use backend_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)
	_ = d.Set("pipeline_id", types.FlattenStringPtr(&stage.PipelineID))

	return setBackendStageState(d, stage, zone)
}
