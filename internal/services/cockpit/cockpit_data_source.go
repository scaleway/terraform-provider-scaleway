package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpit() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCockpit().Schema)

	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ValidateFunc: verify.IsUUID(),
	}
	delete(dsSchema, "plan")

	return &schema.Resource{
		ReadContext: dataSourceCockpitRead,
		Schema:      dsSchema,
	}
}

func dataSourceCockpitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	res, err := waitForCockpit(ctx, api, projectID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ProjectID)
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("plan_id", res.Plan.ID)
	_ = d.Set("endpoints", flattenCockpitEndpoints(res.Endpoints))

	return nil
}
