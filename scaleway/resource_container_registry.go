package scaleway

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

import "github.com/scaleway/scaleway-sdk-go/api/registry/v1"

func resourceScalewayContainerRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayContainerRegistryCreate,
		Read:   resourceScalewayContainerRegistryRead,
		Update: resourceScalewayContainerRegistryUpdate,
		Delete: resourceScalewayContainerRegistryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the container registry",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The description of the container registry",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "Define the default visibity policy",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint reachable by docker",
			},
		},
	}
}

func resourceScalewayContainerRegistryCreate(d *schema.ResourceData, m interface{}) error {
	api, region, err := containerRegistryWithRegion(d, m)
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("cr")
	}

	ns, err := api.CreateNamespace(&registry.CreateNamespaceRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           name.(string),
		Description:    d.Get("description").(string),
		IsPublic:       d.Get("is_public").(bool),
	})
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, ns.ID))

	return resourceScalewayContainerRegistryRead(d, m)
}

func resourceScalewayContainerRegistryRead(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := containerRegistryWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	ns, err := api.GetNamespace(&registry.GetNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("name", ns.Name)
	_ = d.Set("description", ns.Description)
	_ = d.Set("organization_id", ns.OrganizationID)
	_ = d.Set("is_public", ns.IsPublic)
	_ = d.Set("endpoint", ns.Endpoint)

	return nil
}

func resourceScalewayContainerRegistryUpdate(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := containerRegistryWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	description := d.Get("description").(string)
	isPublic := d.Get("is_public").(bool)

	if d.HasChange("description") || d.HasChange("is_public") {
		if _, err := api.UpdateNamespace(&registry.UpdateNamespaceRequest{
			Region:      region,
			NamespaceID: id,
			Description: &description,
			IsPublic:    &isPublic,
		}); err != nil {
			return err
		}
	}

	return resourceScalewayContainerRegistryRead(d, m)
}

func resourceScalewayContainerRegistryDelete(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := containerRegistryWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	_, err = api.DeleteNamespace(&registry.DeleteNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
