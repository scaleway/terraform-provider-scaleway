package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayFunction() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayFunction().Schema)

	dsSchema["function_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the function",
		Computed:    true,
	}

	addOptionalFieldsToSchema(dsSchema, "name", "function_id", "project_id", "region")
	fixDatasourceSchemaFlags(dsSchema, true, "namespace_id")

	return &schema.Resource{
		ReadContext: dataSourceScalewayFunctionRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	functionID, ok := d.GetOk("function_id")
	if !ok {
		functionName := d.Get("name").(string)
		res, err := api.ListFunctions(&function.ListFunctionsRequest{
			Region:      region,
			NamespaceID: expandID(d.Get("namespace_id").(string)),
			Name:        expandStringPtr(functionName),
			ProjectID:   expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundFunction, err := findExact(
			res.Functions,
			func(s *function.Function) bool { return s.Name == functionName },
			functionName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		functionID = foundFunction.ID
	}

	regionalID := datasourceNewRegionalID(functionID, region)
	d.SetId(regionalID)
	_ = d.Set("function_id", regionalID)

	return resourceScalewayFunctionRead(ctx, d, meta)
}
