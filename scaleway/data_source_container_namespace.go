package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayContainerNamespace() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayContainerNamespace().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"namespace_id"}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Container namespace",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayContainerNamespaceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayContainerNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID, ok := d.GetOk("namespace_id")
	if !ok {
		res, err := api.ListNamespaces(&container.ListNamespacesRequest{
			Region: region,
			Name:   expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Namespaces) == 0 {
			return diag.FromErr(fmt.Errorf("no container namespace found with the name %s", d.Get("name")))
		}
		if len(res.Namespaces) > 1 {
			return diag.FromErr(fmt.Errorf("%d container namespaces found with the same name %s", len(res.Namespaces), d.Get("name")))
		}
		namespaceID = res.Namespaces[0].ID
	}

	regionalID := datasourceNewRegionalizedID(namespaceID, region)
	d.SetId(regionalID)
	_ = d.Set("namespace_id", regionalID)

	return resourceScalewayContainerNamespaceRead(ctx, d, meta)
}
