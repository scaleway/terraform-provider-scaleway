package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpit() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCockpit().Schema)
	dsSchema["project_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The project_id you want to attach the resource to",
		Optional:         true,
		ValidateDiagFunc: verify.IsUUID(),
	}
	dsSchema["plan"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "[DEPRECATED] The current plan of the cockpit project.",
		Deprecated:  "The 'plan' attribute is deprecated and will be removed in a future version. Any changes to this attribute will have no effect.",
	}

	return &schema.Resource{
		ReadContext:        dataSourceCockpitRead,
		Schema:             dsSchema,
		DeprecationMessage: "The 'scaleway_cockpit' data source is deprecated because it duplicates the functionality of the 'scaleway_cockpit' resource. Please use the 'scaleway_cockpit' resource instead.",
	}
}

func dataSourceCockpitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectID := d.Get("project_id").(string)
	if projectID == "" {
		_, err := getDefaultProjectID(ctx, m)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := diag.Diagnostics{}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Deprecated attribute: 'plan'",
		Detail:   "The 'plan' attribute is deprecated and will be removed in a future version. Any changes to this attribute will have no effect.",
	})

	_ = d.Set("plan", d.Get("plan"))
	_ = d.Set("project_id", projectID)
	d.SetId(projectID)

	return diags
}
