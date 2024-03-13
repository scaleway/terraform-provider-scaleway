package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayRegistryNamespace() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRegistryNamespace().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"namespace_id"}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the registry namespace",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
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
		namespaceName := d.Get("name").(string)
		res, err := api.ListNamespaces(&registry.ListNamespacesRequest{
			Region:    region,
			Name:      expandStringPtr(namespaceName),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundNamespace, err := findExact(
			res.Namespaces,
			func(s *registry.Namespace) bool { return s.Name == namespaceName },
			namespaceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceID = foundNamespace.ID
	}

	regionalID := datasourceNewRegionalID(namespaceID, region)
	d.SetId(regionalID)
	_ = d.Set("namespace_id", regionalID)

	return resourceScalewayRegistryNamespaceRead(ctx, d, m)
}
