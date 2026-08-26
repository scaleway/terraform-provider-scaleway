package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstancePolicy() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstancePolicy().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "instance_group_id", "zone")

	dsSchema["name"].ConflictsWith = []string{"instance_policy_id"}
	dsSchema["instance_policy_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the instance policy",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext:        DataSourceInstancePolicyRead,
		Schema:             dsSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func DataSourceInstancePolicyRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_autoscaling_instance_policy data source is no longer supported",
		Detail:   deprecationMessage,
	}}
}
