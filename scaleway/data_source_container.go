package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

	return &schema.Resource{
		ReadContext: dataSourceScalewayContainerRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	containerID, ok := d.GetOk("container_id")
	namespaceID := d.Get("namespace_id")
	if !ok {
		res, err := api.ListContainers(&container.ListContainersRequest{
			Region:      region,
			Name:        expandStringPtr(d.Get("name")),
			NamespaceID: expandID(namespaceID),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Containers) == 0 {
			return diag.FromErr(fmt.Errorf("no container found with the name %s", d.Get("name")))
		}
		if len(res.Containers) > 1 {
			return diag.FromErr(fmt.Errorf("%d container found with the same name %s", len(res.Containers), d.Get("name")))
		}
		containerID = res.Containers[0].ID
	}

	regionalID := datasourceNewRegionalizedID(containerID, region)
	d.SetId(regionalID)
	_ = d.Set("container_id", regionalID)

	return resourceScalewayContainerRead(ctx, d, meta)
}
