package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRegistryNamespace() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRegistryNamespace().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"namespace_id"}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the registry namespace",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayRegistryNamespaceRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayRegistryNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := registryAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID, ok := d.GetOk("namespace_id")
	if !ok {
		res, err := api.ListNamespaces(&registry.ListNamespacesRequest{
			Region: region,
			Name:   expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Namespaces) == 0 {
			return diag.FromErr(fmt.Errorf("no namespaces found with the name %s", d.Get("name")))
		}
		if len(res.Namespaces) > 1 {
			return diag.FromErr(fmt.Errorf("%d namespaces found with the same name %s", len(res.Namespaces), d.Get("name")))
		}
		namespaceID = res.Namespaces[0].ID
	}

	regionalID := datasourceNewRegionalizedID(namespaceID, region)
	d.SetId(regionalID)
	_ = d.Set("namespace_id", regionalID)

	return resourceScalewayRegistryNamespaceRead(ctx, d, m)
}
