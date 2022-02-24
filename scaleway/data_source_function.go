package scaleway

import (
	"context"
	"fmt"

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
		res, err := api.ListFunctions(&function.ListFunctionsRequest{
			Region: region,
			Name:   expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Functions) == 0 {
			return diag.FromErr(fmt.Errorf("no functions found with the name %s", d.Get("name")))
		}
		if len(res.Functions) > 1 {
			return diag.FromErr(fmt.Errorf("%d functions found with the same name %s", len(res.Functions), d.Get("name")))
		}
		functionID = res.Functions[0].ID
	}

	regionalID := datasourceNewRegionalizedID(functionID, region)
	d.SetId(regionalID)
	_ = d.Set("function_id", regionalID)

	return resourceScalewayFunctionRead(ctx, d, meta)
}
