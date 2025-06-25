package container

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceNamespace() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceNamespace().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"namespace_id"}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the Container namespace",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceContainerNamespaceRead,
		Schema:      dsSchema,
	}
}

func DataSourceContainerNamespaceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID, ok := d.GetOk("namespace_id")
	if !ok {
		namespaceName := d.Get("name").(string)

		res, err := api.ListNamespaces(&container.ListNamespacesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(namespaceName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundNamespace, err := datasource.FindExact(
			res.Namespaces,
			func(s *container.Namespace) bool { return s.Name == namespaceName },
			namespaceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceID = foundNamespace.ID
	}

	regionalID := datasource.NewRegionalID(namespaceID, region)
	d.SetId(regionalID)
	_ = d.Set("namespace_id", regionalID)

	return ResourceContainerNamespaceRead(ctx, d, m)
}
