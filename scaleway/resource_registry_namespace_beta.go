package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRegistryNamespaceBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayRegistryNamespaceBetaCreate,
		Read:   resourceScalewayRegistryNamespaceBetaRead,
		Update: resourceScalewayRegistryNamespaceBetaUpdate,
		Delete: resourceScalewayRegistryNamespaceBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the container registry namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the container registry namespace",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Define the default visibity policy",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint reachable by docker",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayRegistryNamespaceBetaCreate(d *schema.ResourceData, m interface{}) error {
	api, region, err := registryNamespaceAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	ns, err := api.CreateNamespace(&registry.CreateNamespaceRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		IsPublic:       d.Get("is_public").(bool),
	})
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, ns.ID))

	return resourceScalewayRegistryNamespaceBetaRead(d, m)
}

func resourceScalewayRegistryNamespaceBetaRead(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := registryNamespaceAPIWithRegionAndID(m, d.Id())
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

	d.Set("name", ns.Name)
	d.Set("description", ns.Description)
	d.Set("organization_id", ns.OrganizationID)
	d.Set("is_public", ns.IsPublic)
	d.Set("endpoint", ns.Endpoint)
	d.Set("region", ns.Region)

	return nil
}

func resourceScalewayRegistryNamespaceBetaUpdate(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := registryNamespaceAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("description") || d.HasChange("is_public") {
		if _, err := api.UpdateNamespace(&registry.UpdateNamespaceRequest{
			Region:      region,
			NamespaceID: id,
			Description: scw.StringPtr(d.Get("description").(string)),
			IsPublic:    scw.BoolPtr(d.Get("is_public").(bool)),
		}); err != nil {
			return err
		}
	}

	return resourceScalewayRegistryNamespaceBetaRead(d, m)
}

func resourceScalewayRegistryNamespaceBetaDelete(d *schema.ResourceData, m interface{}) error {
	api, region, id, err := registryNamespaceAPIWithRegionAndID(m, d.Id())
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
