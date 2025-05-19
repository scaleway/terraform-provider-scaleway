package container

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	registrySDK "github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceContainerNamespaceCreate,
		ReadContext:   ResourceContainerNamespaceRead,
		UpdateContext: ResourceContainerNamespaceUpdate,
		DeleteContext: ResourceContainerNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultContainerNamespaceTimeout),
			Read:    schema.DefaultTimeout(defaultContainerNamespaceTimeout),
			Update:  schema.DefaultTimeout(defaultContainerNamespaceTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerNamespaceTimeout),
			Default: schema.DefaultTimeout(defaultContainerNamespaceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
				Description: "The name of the container namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the container namespace",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the container namespace",
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The environment variables of the container namespace",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
			},
			"secret_environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Description: "The secret environment variables of the container namespace",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc:      validation.MapKeyLenBetween(0, 100),
				DiffSuppressFunc:      dsf.CompareArgon2idPasswordAndHash,
				DiffSuppressOnRefresh: true,
			},
			"registry_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint reachable by docker",
			},
			"registry_namespace_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the registry namespace",
			},
			"destroy_registry": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Destroy registry on deletion",
				Deprecated:  "Registry namespace is automatically destroyed with namespace",
			},
			"region":          regional.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func ResourceContainerNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectId, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &container.CreateNamespaceRequest{
		Description:                types.ExpandStringPtr(d.Get("description").(string)),
		EnvironmentVariables:       types.ExpandMapPtrStringString(d.Get("environment_variables")),
		SecretEnvironmentVariables: expandContainerSecrets(d.Get("secret_environment_variables")),
		Name:                       types.ExpandOrGenerateString(d.Get("name").(string), "ns"),
		ProjectID:                  projectId,
		Region:                     region,
	}

	rawTag, tagExist := d.GetOk("tags")
	if tagExist {
		createReq.Tags = types.ExpandStrings(rawTag)
	}

	ns, err := api.CreateNamespace(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, ns.ID))

	_, err = waitForNamespace(ctx, api, region, ns.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerNamespaceRead(ctx, d, m)
}

func ResourceContainerNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("description", types.FlattenStringPtr(ns.Description))
	_ = d.Set("tags", types.FlattenSliceString(ns.Tags))
	_ = d.Set("environment_variables", ns.EnvironmentVariables)
	_ = d.Set("name", ns.Name)
	_ = d.Set("organization_id", ns.OrganizationID)
	_ = d.Set("project_id", ns.ProjectID)
	_ = d.Set("region", ns.Region)
	_ = d.Set("registry_endpoint", ns.RegistryEndpoint)
	_ = d.Set("registry_namespace_id", ns.RegistryNamespaceID)
	_ = d.Set("secret_environment_variables", flattenContainerSecrets(ns.SecretEnvironmentVariables))

	return nil
}

func ResourceContainerNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.UpdateNamespaceRequest{
		Region:      ns.Region,
		NamespaceID: ns.ID,
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChanges("environment_variables") {
		req.EnvironmentVariables = types.ExpandMapPtrStringString(d.Get("environment_variables"))
	}

	if d.HasChanges("secret_environment_variables") {
		oldEnv, newEnv := d.GetChange("secret_environment_variables")
		req.SecretEnvironmentVariables = filterSecretEnvsToPatch(expandContainerSecrets(oldEnv), expandContainerSecrets(newEnv))
	}

	if _, err := api.UpdateNamespace(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerNamespaceRead(ctx, d, m)
}

func ResourceContainerNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_, err = api.DeleteNamespace(&container.DeleteNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	if destroy := d.Get("destroy_registry"); destroy != nil && destroy == true {
		registryAPI, region, err := registry.NewAPIWithRegion(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		registryID := d.Get("registry_namespace_id").(string)

		_, err = registryAPI.DeleteNamespace(&registrySDK.DeleteNamespaceRequest{
			Region:      region,
			NamespaceID: registryID,
		})
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}

		_, err = registry.WaitForNamespace(ctx, registryAPI, region, registryID, d.Timeout(schema.TimeoutDelete))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
