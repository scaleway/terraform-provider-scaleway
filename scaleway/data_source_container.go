package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayContainer() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayContainer().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"container_id"}
	dsSchema["container_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Container",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["namespace_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The ID of the Container namespace",
		ValidateFunc: validationUUIDorUUIDWithLocality(),
	}
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the project to filter the Container",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayContainerRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	containerID, ok := d.GetOk("container_id")
	namespaceID := d.Get("namespace_id")
	if !ok {
		containerName := d.Get("name").(string)
		res, err := api.ListContainers(&container.ListContainersRequest{
			Region:      region,
			Name:        expandStringPtr(containerName),
			NamespaceID: locality.ExpandID(namespaceID),
			ProjectID:   expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundContainer, err := findExact(
			res.Containers,
			func(s *container.Container) bool { return s.Name == containerName },
			containerName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		containerID = foundContainer.ID
	}

	regionalID := datasourceNewRegionalID(containerID, region)
	d.SetId(regionalID)
	_ = d.Set("container_id", regionalID)

	return resourceScalewayContainerRead(ctx, d, m)
}
