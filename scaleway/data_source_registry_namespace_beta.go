package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
)

func dataSourceScalewayRegistryNamespace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayRegistryNamespaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the Registry Namespace",
				ConflictsWith: []string{"namespace_id"},
			},
			"namespace_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the Registry Namespace",
				ConflictsWith: []string{"name"},
				ValidateFunc:  validationUUIDorUUIDWithLocality(),
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint of the Registry Namespace",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The Registry Namespace visibility policy",
			},
		},
	}
}

func dataSourceScalewayRegistryNamespaceRead(d *schema.ResourceData, m interface{}) error {
	api, region, err := registryNamespaceAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	var namespace *registry.Namespace
	namespaceID, ok := d.GetOk("namespace_id")
	if !ok {
		res, err := api.ListNamespaces(&registry.ListNamespacesRequest{
			Region: region,
			Name:   String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.Namespaces) == 0 {
			return fmt.Errorf("no namespaces found with the name %s", d.Get("name"))
		}
		if len(res.Namespaces) > 1 {
			return fmt.Errorf("%d namespaces found with the same name %s", len(res.Namespaces), d.Get("name"))
		}
		namespace = res.Namespaces[0]
	} else {
		res, err := api.GetNamespace(&registry.GetNamespaceRequest{
			Region:      region,
			NamespaceID: expandID(namespaceID),
		})
		if err != nil {
			return err
		}
		namespace = res
	}

	d.SetId(datasourceNewRegionalID(namespace.ID, region))
	_ = d.Set("namespace_id", namespace.ID)
	_ = d.Set("name", namespace.Name)
	_ = d.Set("endpoint", namespace.Endpoint)
	_ = d.Set("is_public", namespace.IsPublic)

	return nil
}
