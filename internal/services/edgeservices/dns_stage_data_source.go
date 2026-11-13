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

//go:embed descriptions/dns_stage_data_source.md
var dnsStageDataSourceDescription string

func DataSourceDNSStage() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDNSStage().SchemaFunc())

	filterFields := []string{"pipeline_id", "fqdn"}

	dsSchema["dns_stage_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the DNS stage",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    filterFields,
	}

	dsSchema["fqdn"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "FQDN to filter for (in the format subdomain.example.com)",
		ConflictsWith: []string{"dns_stage_id"},
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "pipeline_id")

	dsSchema["pipeline_id"].ConflictsWith = []string{"dns_stage_id"}

	return &schema.Resource{
		ReadContext: DataSourceDNSStageRead,
		Description: dnsStageDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceDNSStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	stageID, idExists := d.GetOk("dns_stage_id")
	if idExists {
		return dataSourceDNSStageReadByID(ctx, d, m, stageID.(string))
	}

	return dataSourceDNSStageReadByFilters(ctx, d, m)
}

func dataSourceDNSStageReadByID(ctx context.Context, d *schema.ResourceData, m any, stageID string) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	d.SetId(stageID)

	dnsStage, err := api.GetDNSStage(&edgeservices.GetDNSStageRequest{
		DNSStageID: stageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setDNSStageState(d, dnsStage)
}

func dataSourceDNSStageReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	pipelineID, ok := d.GetOk("pipeline_id")
	if !ok {
		return diag.FromErr(errors.New("pipeline_id is required when dns_stage_id is not specified"))
	}

	req := &edgeservices.ListDNSStagesRequest{
		PipelineID: pipelineID.(string),
		Fqdn:       types.ExpandStringPtr(d.Get("fqdn")),
	}

	res, err := api.ListDNSStages(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Stages) == 0 {
		return diag.FromErr(errors.New("no DNS stage found matching the specified filters"))
	}

	if len(res.Stages) > 1 {
		return diag.FromErr(fmt.Errorf("multiple DNS stages (%d) found, please refine your filters or use dns_stage_id", len(res.Stages)))
	}

	stage := res.Stages[0]
	d.SetId(stage.ID)

	return setDNSStageState(d, stage)
}
