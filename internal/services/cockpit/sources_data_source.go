package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceCockpitSources() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCockpitSource().SchemaFunc())

	sourceElementSchema := datasource.SchemaFromResourceSchema(ResourceCockpitSource().SchemaFunc())
	delete(sourceElementSchema, "sources")

	sourceElementSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the data source.",
		Computed:    true,
	}

	dsSchema["sources"] = &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of cockpit sources.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: sourceElementSchema,
		},
	}

	datasource.FixDatasourceSchemaFlags(dsSchema, false, "name", "project_id")
	datasource.AddOptionalFieldsToSchema(dsSchema, "region", "type", "origin")

	return &schema.Resource{
		ReadContext: dataSourceCockpitSourcesRead,
		Schema:      dsSchema,
	}
}

func dataSourceCockpitSourcesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var region scw.Region

	var err error

	if v, ok := d.GetOk("region"); ok && v.(string) != "" {
		region, err = scw.ParseRegion(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		_, region, err = cockpitAPIWithRegion(d, m)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	api := cockpit.NewRegionalAPI(meta.ExtractScwClient(m))

	req := &cockpit.RegionalAPIListDataSourcesRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
	}

	if v, ok := d.GetOk("type"); ok {
		req.Types = []cockpit.DataSourceType{cockpit.DataSourceType(v.(string))}
	}

	if v, ok := d.GetOk("origin"); ok {
		req.Origin = cockpit.DataSourceOrigin(v.(string))
	}

	res, err := api.ListDataSources(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	sources := []any(nil)

	for _, ds := range res.DataSources {
		if name, ok := d.GetOk("name"); ok {
			if ds.Name != name.(string) {
				continue
			}
		}

		rawSource := flattenDataSourceToMap(ds)
		sources = append(sources, rawSource)
	}

	d.SetId(region.String())
	_ = d.Set("sources", sources)

	return nil
}

func flattenDataSourceToMap(ds *cockpit.DataSource) map[string]any {
	pushURL, _ := createCockpitPushURL(ds.Type, ds.URL)

	return map[string]any{
		"id":                        regional.NewIDString(ds.Region, ds.ID),
		"project_id":                ds.ProjectID,
		"name":                      ds.Name,
		"url":                       ds.URL,
		"type":                      ds.Type.String(),
		"origin":                    ds.Origin.String(),
		"created_at":                types.FlattenTime(ds.CreatedAt),
		"updated_at":                types.FlattenTime(ds.UpdatedAt),
		"synchronized_with_grafana": ds.SynchronizedWithGrafana,
		"retention_days":            int(ds.RetentionDays),
		"region":                    ds.Region.String(),
		"push_url":                  pushURL,
	}
}
