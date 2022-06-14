package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayFunctionNamespace() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayFunctionNamespace().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"namespace_id"}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the function namespace",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayFunctionNamespaceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayFunctionNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID, ok := d.GetOk("namespace_id")
	if !ok {
		res, err := api.ListNamespaces(&function.ListNamespacesRequest{
			Region: region,
			Name:   expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Namespaces) == 0 {
			return diag.FromErr(fmt.Errorf("no function namespaces found with the name %s", d.Get("name")))
		}
		if len(res.Namespaces) > 1 {
			return diag.FromErr(fmt.Errorf("%d function namespaces found with the same name %s", len(res.Namespaces), d.Get("name")))
		}
		namespaceID = res.Namespaces[0].ID
	}

	regionalID := datasourceNewRegionalizedID(namespaceID, region)
	d.SetId(regionalID)
	_ = d.Set("namespace_id", regionalID)

	return resourceScalewayFunctionNamespaceRead(ctx, d, meta)
}
