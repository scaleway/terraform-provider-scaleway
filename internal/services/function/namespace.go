package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFunctionNamespaceCreate,
		ReadContext:   ResourceFunctionNamespaceRead,
		UpdateContext: ResourceFunctionNamespaceUpdate,
		DeleteContext: ResourceFunctionNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultFunctionNamespaceTimeout),
			Read:    schema.DefaultTimeout(defaultFunctionNamespaceTimeout),
			Update:  schema.DefaultTimeout(defaultFunctionNamespaceTimeout),
			Delete:  schema.DefaultTimeout(defaultFunctionNamespaceTimeout),
			Default: schema.DefaultTimeout(defaultFunctionNamespaceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
				Description: "The name of the function namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the function namespace",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the function namespace",
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The environment variables of the function namespace",
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
				Description: "The environment variables of the function namespace",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
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
			"region":          regional.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func ResourceFunctionNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &function.CreateNamespaceRequest{
		Description:                types.ExpandStringPtr(d.Get("description").(string)),
		EnvironmentVariables:       types.ExpandMapPtrStringString(d.Get("environment_variables")),
		SecretEnvironmentVariables: expandFunctionsSecrets(d.Get("secret_environment_variables")),
		Name:                       types.ExpandOrGenerateString(d.Get("name").(string), "func"),
		ProjectID:                  d.Get("project_id").(string),
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

	return ResourceFunctionNamespaceRead(ctx, d, m)
}

func ResourceFunctionNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	_ = d.Set("description", ns.Description)
	_ = d.Set("tags", types.FlattenSliceString(ns.Tags))
	_ = d.Set("environment_variables", ns.EnvironmentVariables)
	_ = d.Set("name", ns.Name)
	_ = d.Set("organization_id", ns.OrganizationID)
	_ = d.Set("project_id", ns.ProjectID)
	_ = d.Set("region", ns.Region)
	_ = d.Set("registry_endpoint", ns.RegistryEndpoint)
	_ = d.Set("registry_namespace_id", ns.RegistryNamespaceID)

	return nil
}

func ResourceFunctionNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &function.UpdateNamespaceRequest{
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
		req.SecretEnvironmentVariables = expandFunctionsSecrets(d.Get("secret_environment_variables"))
	}

	if _, err := api.UpdateNamespace(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceFunctionNamespaceRead(ctx, d, m)
}

func ResourceFunctionNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteNamespace(&function.DeleteNamespaceRequest{
		Region:      region,
		NamespaceID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForNamespace(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
