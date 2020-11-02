package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRegistryNamespaceBeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRegistryNamespaceBetaCreate,
		ReadContext:   resourceScalewayRegistryNamespaceBetaRead,
		UpdateContext: resourceScalewayRegistryNamespaceBetaUpdate,
		DeleteContext: resourceScalewayRegistryNamespaceBetaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceScalewayRegistryNamespaceBetaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := registryAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := api.CreateNamespace(&registry.CreateNamespaceRequest{
		Region: region,
		// Registry does not support yet project id.
		// Once it does, use ProjectID instead or OrganizationID
		OrganizationID: expandStringPtr(d.Get("project_id")),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		IsPublic:       d.Get("is_public").(bool),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, ns.ID))

	return resourceScalewayRegistryNamespaceBetaRead(ctx, d, m)
}

func resourceScalewayRegistryNamespaceBetaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := registryAPIWithRegionAndID(m, d.Id())
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
	// Registry does not support yet project id.
	// Once it does, use ProjectID instead or OrganizationID
	_ = d.Set("project_id", ns.OrganizationID)
	_ = d.Set("is_public", ns.IsPublic)
	_ = d.Set("endpoint", ns.Endpoint)
	_ = d.Set("region", ns.Region)

	return nil
}

func resourceScalewayRegistryNamespaceBetaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := registryAPIWithRegionAndID(m, d.Id())
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

	return resourceScalewayRegistryNamespaceBetaRead(ctx, d, m)
}

func resourceScalewayRegistryNamespaceBetaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := registryAPIWithRegionAndID(m, d.Id())
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
