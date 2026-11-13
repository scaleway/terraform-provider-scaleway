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

//go:embed descriptions/tls_stage_data_source.md
var tlsStageDataSourceDescription string

func DataSourceTLSStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceTLSStage().SchemaFunc())

	filterFields := []string{"pipeline_id", "secret_id", "secret_region"}

	dsSchema["tls_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the TLS stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}
	dsSchema["secret_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Secret ID to filter for. Only TLS stages with this Secret ID will be returned",
		ConflictsWith: []string{"tls_stage_id"},
	}
	dsSchema["secret_region"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Secret region to filter for. Only TLS stages with a Secret in this region will be returned",
		ConflictsWith: []string{"tls_stage_id"},
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")

	dsSchema["pipeline_id"].ConflictsWith = []string{"tls_stage_id"}

	return &schema.Resource{
		ReadContext: DataSourceTLSStageRead,
		Description: tlsStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceTLSStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("tls_stage_id")
	if idExists {
		return dataSourceTLSStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceTLSStageReadByFilters(ctx, d, m)
}

func dataSourceTLSStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(stageID)

	tlsStage, err := api.GetTLSStage(&edgeservices.GetTLSStageRequest{
		TLSStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setTLSStageState(d, tlsStage)
}

func dataSourceTLSStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when tls_stage_id is not specified"))
	}

	req := &edgeservices.ListTLSStagesRequest{
		PipelineID:   pipelineID.(string),
		SecretID:     types.ExpandStringPtr(d.Get("secret_id")),
		SecretRegion: types.ExpandStringPtr(d.Get("secret_region")),
	}

	res, err := api.ListTLSStages(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no TLS stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple TLS stages (%d) found, please refine your filters or use tls_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)

	return setTLSStageState(d, stage)
}
