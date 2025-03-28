package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceCockpitPlanRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "[DEPRECATED] The plan field is deprecated.",
				Deprecated:  "The 'plan' attribute is deprecated and no longer has any effect. Future updates will remove this attribute entirely.",
			},
		},
		DeprecationMessage: "This data source is deprecated and will be removed in the next major version. Use `my_new_data_source` instead.",
	}
}

func DataSourceCockpitPlanRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("free")
	_ = d.Set("name", "free")

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Data source 'cockpit_plan' has been removed",
			Detail:   "The 'cockpit_plan' data source has been deprecated and is no longer available.",
		},
	}
}
