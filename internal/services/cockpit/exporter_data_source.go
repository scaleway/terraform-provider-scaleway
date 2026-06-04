package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceCockpitExporter() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(exporterSchema())

	dsSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "ID of the exporter.",
	}

	datasource.FixDatasourceSchemaFlags(dsSchema, false, "id", "name", "project_id")
	datasource.AddOptionalFieldsToSchema(dsSchema, "region")

	return &schema.Resource{
		ReadContext: dataSourceCockpitExporterRead,
		Schema:      dsSchema,
	}
}

func dataSourceCockpitExporterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, ok := d.GetOk("id"); ok {
		return fetchExporterByID(ctx, d, meta)
	}

	return fetchExporterByFilters(ctx, d, meta)
}

func fetchExporterByID(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetExporter(&cockpit.RegionalAPIGetExporterRequest{
		Region:     region,
		ExporterID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	flattenExporterWithProjectID(ctx, d, api, region, res)

	return nil
}

func fetchExporterByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id is required when not fetching by id")
	}

	req := &cockpit.RegionalAPIListExportersRequest{
		Region:    region,
		ProjectID: projectID,
	}

	res, err := api.ListExporters(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	if res.TotalCount == 0 {
		return diag.Errorf("no exporter found matching the specified criteria")
	}

	if name, ok := d.GetOk("name"); ok {
		for _, exp := range res.Exporters {
			if exp.Name == name.(string) {
				flattenExporterWithProjectID(ctx, d, api, region, exp)

				return nil
			}
		}

		return diag.Errorf("no exporter found with name '%s'", name.(string))
	}

	flattenExporterWithProjectID(ctx, d, api, region, res.Exporters[0])

	return nil
}

func flattenExporterWithProjectID(ctx context.Context, d *schema.ResourceData, api *cockpit.RegionalAPI, region scw.Region, exp *cockpit.Exporter) {
	d.SetId(regional.NewIDString(region, exp.ID))
	_ = d.Set("name", exp.Name)
	_ = d.Set("description", exp.Description)
	_ = d.Set("datasource_id", regional.NewIDString(region, exp.DatasourceID))
	_ = d.Set("status", exp.Status.String())
	_ = d.Set("region", region.String())
	_ = d.Set("created_at", types.FlattenTime(exp.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(exp.UpdatedAt))
	_ = d.Set("exported_products", types.FlattenSliceString(exp.ExportedProducts))

	ds, err := api.GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
		Region:       region,
		DataSourceID: exp.DatasourceID,
	}, scw.WithContext(ctx))
	if err == nil {
		_ = d.Set("project_id", ds.ProjectID)
	}

	if exp.DatadogDestination != nil {
		datadogDest := map[string]any{"endpoint": ""}
		if exp.DatadogDestination.Endpoint != nil {
			datadogDest["endpoint"] = *exp.DatadogDestination.Endpoint
		}

		_ = d.Set("datadog_destination", []map[string]any{datadogDest})
	}

	if exp.OtlpDestination != nil {
		otlpDest := []map[string]any{
			{
				"endpoint": exp.OtlpDestination.Endpoint,
				"headers":  types.FlattenMap(exp.OtlpDestination.Headers),
			},
		}
		_ = d.Set("otlp_destination", otlpDest)
	}
}
