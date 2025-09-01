package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceFunction() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceFunction().Schema)

	dsSchema["function_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the function",
		Computed:    true,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "function_id", "project_id", "region")
	datasource.FixDatasourceSchemaFlags(dsSchema, true, "namespace_id")

	return &schema.Resource{
		ReadContext: DataSourceFunctionRead,
		Schema:      dsSchema,
	}
}

func DataSourceFunctionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	functionID, ok := d.GetOk("function_id")
	if !ok {
		functionName := d.Get("name").(string)

		res, err := api.ListFunctions(&function.ListFunctionsRequest{
			Region:      region,
			NamespaceID: locality.ExpandID(d.Get("namespace_id").(string)),
			Name:        types.ExpandStringPtr(functionName),
			ProjectID:   types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundFunction, err := datasource.FindExact(
			res.Functions,
			func(s *function.Function) bool { return s.Name == functionName },
			functionName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		functionID = foundFunction.ID
	}

	regionalID := datasource.NewRegionalID(functionID, region)
	d.SetId(regionalID)
	_ = d.Set("function_id", regionalID)

	return ResourceFunctionRead(ctx, d, m)
}
