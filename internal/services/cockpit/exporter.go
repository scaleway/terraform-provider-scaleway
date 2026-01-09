package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceCockpitExporter() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitExporterCreate,
		ReadContext:   ResourceCockpitExporterRead,
		UpdateContext: ResourceCockpitExporterUpdate,
		DeleteContext: ResourceCockpitExporterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Update:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: exporterSchema,
	}
}

func exporterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the data export",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Description of the data export",
		},
		"datasource_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "ID of the data source linked to the data export",
			DiffSuppressFunc: dsf.Locality,
		},
		"datadog_destination": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Datadog destination configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"api_key": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "Datadog API key",
					},
					"endpoint": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Datadog endpoint URL",
					},
				},
			},
		},
		"otlp_destination": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "OTLP destination configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "OTLP endpoint URL",
					},
					"headers": {
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Headers to include in requests",
					},
				},
			},
		},
		"exported_products": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of Scaleway products available for metrics/logs data exports",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the data export",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of the creation of the data export (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of the last update of the data export (RFC 3339 format)",
		},
		"project_id": account.ProjectIDSchema(),
		"region":     regional.Schema(),
	}
}

func ResourceCockpitExporterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	datasourceID := locality.ExpandID(d.Get("datasource_id").(string))

	req := &cockpit.RegionalAPICreateExporterRequest{
		Region:       region,
		DatasourceID: datasourceID,
		Name:         d.Get("name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = types.ExpandStringPtr(v)
	}

	if v, ok := d.GetOk("exported_products"); ok {
		req.ExportedProducts = types.ExpandStrings(v)
	}

	datadogDest, hasDatadog := d.GetOk("datadog_destination")
	otlpDest, hasOTLP := d.GetOk("otlp_destination")

	if hasDatadog && hasOTLP {
		return diag.Errorf("cannot specify both datadog_destination and otlp_destination")
	}

	if !hasDatadog && !hasOTLP {
		return diag.Errorf("must specify either datadog_destination or otlp_destination")
	}

	if hasDatadog {
		datadogList := datadogDest.([]any)
		if len(datadogList) > 0 {
			datadogMap := datadogList[0].(map[string]any)
			req.DatadogDestination = &cockpit.ExporterDatadogDestination{
				APIKey: scw.StringPtr(datadogMap["api_key"].(string)),
			}
			if endpoint, ok := datadogMap["endpoint"]; ok && endpoint != "" {
				req.DatadogDestination.Endpoint = scw.StringPtr(endpoint.(string))
			}
		}
	}

	if hasOTLP {
		otlpList := otlpDest.([]any)
		if len(otlpList) > 0 {
			otlpMap := otlpList[0].(map[string]any)
			req.OtlpDestination = &cockpit.ExporterOTLPDestination{
				Endpoint: otlpMap["endpoint"].(string),
			}
			if headers, ok := otlpMap["headers"]; ok {
				headersMap := types.ExpandMapPtrStringString(headers)
				if headersMap != nil {
					req.OtlpDestination.Headers = *headersMap
				}
			}
		}
	}

	res, err := api.CreateExporter(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	exporter, err := waitForExporter(ctx, api, region, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, exporter.ID))

	return ResourceCockpitExporterRead(ctx, d, meta)
}

func ResourceCockpitExporterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetExporter(&cockpit.RegionalAPIGetExporterRequest{
		Region:     region,
		ExporterID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("description", res.Description)
	_ = d.Set("status", res.Status.String())
	_ = d.Set("region", string(region))
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("exported_products", types.FlattenSliceString(res.ExportedProducts))

	if res.DatadogDestination != nil {
		datadogDest := map[string]any{
			"endpoint": "",
		}
		if res.DatadogDestination.Endpoint != nil {
			datadogDest["endpoint"] = *res.DatadogDestination.Endpoint
		}
		if apiKey, ok := d.GetOk("datadog_destination.0.api_key"); ok {
			datadogDest["api_key"] = apiKey.(string)
		}
		_ = d.Set("datadog_destination", []map[string]any{datadogDest})
	}

	if res.OtlpDestination != nil {
		otlpDest := []map[string]any{
			{
				"endpoint": res.OtlpDestination.Endpoint,
				"headers":  types.FlattenMap(res.OtlpDestination.Headers),
			},
		}
		_ = d.Set("otlp_destination", otlpDest)
	}

	return nil
}

func ResourceCockpitExporterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForExporter(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &cockpit.RegionalAPIUpdateExporterRequest{
		Region:     region,
		ExporterID: id,
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("description") {
		req.Description = types.ExpandStringPtr(d.Get("description"))
	}

	if d.HasChange("exported_products") {
		req.ExportedProducts = types.ExpandStringsPtr(d.Get("exported_products"))
	}

	datadogDest, hasDatadog := d.GetOk("datadog_destination")
	otlpDest, hasOTLP := d.GetOk("otlp_destination")

	if d.HasChange("datadog_destination") || d.HasChange("otlp_destination") {
		if hasDatadog && hasOTLP {
			return diag.Errorf("cannot specify both datadog_destination and otlp_destination")
		}

		if !hasDatadog && !hasOTLP {
			return diag.Errorf("must specify either datadog_destination or otlp_destination")
		}

		if hasDatadog {
			datadogList := datadogDest.([]any)
			if len(datadogList) > 0 {
				datadogMap := datadogList[0].(map[string]any)
				req.DatadogDestination = &cockpit.ExporterDatadogDestination{
					APIKey: scw.StringPtr(datadogMap["api_key"].(string)),
				}
				if endpoint, ok := datadogMap["endpoint"]; ok && endpoint != "" {
					req.DatadogDestination.Endpoint = scw.StringPtr(endpoint.(string))
				}
			}
		}

		if hasOTLP {
			otlpList := otlpDest.([]any)
			if len(otlpList) > 0 {
				otlpMap := otlpList[0].(map[string]any)
				req.OtlpDestination = &cockpit.ExporterOTLPDestination{
					Endpoint: otlpMap["endpoint"].(string),
				}
				if headers, ok := otlpMap["headers"]; ok {
					headersMap := types.ExpandMapPtrStringString(headers)
					if headersMap != nil {
						req.OtlpDestination.Headers = *headersMap
					}
				}
			}
		}
	}

	if d.HasChanges("name", "description", "exported_products", "datadog_destination", "otlp_destination") {
		_, err := api.UpdateExporter(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForExporter(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func ResourceCockpitExporterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForExporter(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteExporter(&cockpit.RegionalAPIDeleteExporterRequest{
		Region:     region,
		ExporterID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForExporter(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
