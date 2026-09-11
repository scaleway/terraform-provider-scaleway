package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstanceGroup() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstanceGroup().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"instance_group_id"}
	dsSchema["instance_group_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the instance group",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext:        DataSourceInstanceGroupRead,
		Schema:             dsSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func DataSourceInstanceGroupRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_autoscaling_instance_group data source is no longer supported",
		Detail:   deprecationMessage,
	}}
}
