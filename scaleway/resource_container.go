package scaleway

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)


func resourceScalewayContainer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerCreate,
		ReadContext:   resourceScalewayContainerRead,
		UpdateContext: resourceScalewayContainerUpdate,
		DeleteContext: resourceScalewayContainerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultContainerTimeout),
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
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiHandler, err := newContainerHandler(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("name_space_id")
	// verify name space state
	apiHandler.waitForNameSpace(expandStringPtr(namespaceID))

	c, err := apiHandler.waitForContainerCreation(d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(apiHandler.region, c.ID))

	return resourceScalewayContainerRead(ctx, d, meta)
}

func resourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := api.GetNamespace(&registry.GetNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", ns.Name)
	_ = d.Set("description", ns.Description)
	_ = d.Set("organization_id", ns.OrganizationID)
	_ = d.Set("project_id", ns.ProjectID)
	_ = d.Set("is_public", ns.IsPublic)
	_ = d.Set("endpoint", ns.Endpoint)
	_ = d.Set("region", ns.Region)

	return nil
}

func resourceScalewayContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("description", "is_public") {
		if _, err := api.UpdateNamespace(&registry.UpdateNamespaceRequest{
			Region:      region,
			NamespaceID: id,
			Description: expandStringPtr(d.Get("description")),
			IsPublic:    scw.BoolPtr(d.Get("is_public").(bool)),
		}, scw.WithContext(ctx)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayContainerRead(ctx, d, meta)
}

func resourceScalewayContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteNamespace(&registry.DeleteNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
