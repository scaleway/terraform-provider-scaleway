package function

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFunctionCreate,
		ReadContext:   ResourceFunctionRead,
		UpdateContext: ResourceFunctionUpdate,
		DeleteContext: ResourceFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(DefaultFunctionTimeout),
			Read:    schema.DefaultTimeout(DefaultFunctionTimeout),
			Update:  schema.DefaultTimeout(DefaultFunctionTimeout),
			Delete:  schema.DefaultTimeout(DefaultFunctionTimeout),
			Create:  schema.DefaultTimeout(DefaultFunctionTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"namespace_id": {
				Type:             schema.TypeString,
				Description:      "The namespace ID associated with this function",
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: dsf.Locality,
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				Description:  "The name of the function",
				ValidateFunc: validation.StringLenBetween(1, 34),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the function",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the function.",
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The environment variables of the function",
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
				Description: "The secret environment variables to be injected into your function at runtime.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc:      validation.MapKeyLenBetween(0, 100),
				DiffSuppressFunc:      dsf.CompareArgon2idPasswordAndHash,
				DiffSuppressOnRefresh: true,
			},
			"privacy": {
				Type:             schema.TypeString,
				Description:      "Privacy of the function. Can be either `private` or `public`",
				Required:         true,
				ValidateDiagFunc: verify.ValidateEnum[function.FunctionPrivacy](),
			},
			"runtime": {
				Type:        schema.TypeString,
				Description: "Runtime of the function",
				Required:    true,
			},
			"min_scale": {
				Type:        schema.TypeInt,
				Description: "Minimum replicas for your function, defaults to 0, Note that a function is billed when it gets executed, and using a min_scale greater than 0 will cause your function to run all the time.",
				Optional:    true,
				Default:     0,
			},
			"max_scale": {
				Type:        schema.TypeInt,
				Description: "Maximum replicas for your function (defaults to 20), our system will scale your functions automatically based on incoming workload, but will never scale the number of replicas above the configured max_scale.",
				Optional:    true,
				Default:     20,
			},
			"memory_limit": {
				Type:        schema.TypeInt,
				Description: "Memory limit in MB for your function, defaults to 256MB",
				Optional:    true,
				Default:     256,
			},
			"handler": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Handler of the function. Depends on the runtime https://developers.scaleway.com/en/products/functions/api/#create-a-function",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Holds the max duration (in seconds) the function is allowed for responding to a request",
				Optional:    true,
			},
			"zip_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location of the zip file to upload containing your function sources",
			},
			"zip_hash": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"zip_file"},
				Description:  "The hash of your source zip file, changing it will re-apply function. Can be any string",
			},
			"deploy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Define if the function should be deployed, terraform will wait for function to be deployed",
			},
			"http_option": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "HTTP traffic configuration",
				Default:          function.FunctionHTTPOptionEnabled.String(),
				ValidateDiagFunc: verify.ValidateEnum[function.FunctionHTTPOption](),
			},
			"sandbox": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Execution environment of the function.",
				ValidateDiagFunc: verify.ValidateEnum[function.FunctionSandbox](),
			},
			"cpu_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU limit in mCPU for your function",
			},
			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The native function domain name.",
			},
			"private_network_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the Private Network the container is connected to",
			},
			"region":          regional.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
		CustomizeDiff: cdf.LocalityCheck("namespace_id"),
	}
}

func ResourceFunctionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, namespace, err := regional.ParseID(d.Get("namespace_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &function.CreateFunctionRequest{
		Description:                types.ExpandStringPtr(d.Get("description").(string)),
		EnvironmentVariables:       types.ExpandMapPtrStringString(d.Get("environment_variables")),
		SecretEnvironmentVariables: expandFunctionsSecrets(d.Get("secret_environment_variables")),
		Handler:                    types.ExpandStringPtr(d.Get("handler").(string)),
		MaxScale:                   types.ExpandUint32Ptr(d.Get("max_scale")),
		MemoryLimit:                types.ExpandUint32Ptr(d.Get("memory_limit")),
		MinScale:                   types.ExpandUint32Ptr(d.Get("min_scale")),
		Name:                       types.ExpandOrGenerateString(d.Get("name").(string), "func"),
		NamespaceID:                namespace,
		Privacy:                    function.FunctionPrivacy(d.Get("privacy").(string)),
		Region:                     region,
		Runtime:                    function.FunctionRuntime(d.Get("runtime").(string)),
		HTTPOption:                 function.FunctionHTTPOption(d.Get("http_option").(string)),
		Sandbox:                    function.FunctionSandbox(d.Get("sandbox").(string)),
	}

	if tags, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(tags)
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		req.Timeout = &scw.Duration{Seconds: int64(timeout.(int))}
	}

	if pnID, ok := d.GetOk("private_network_id"); ok {
		req.PrivateNetworkID = types.ExpandStringPtr(locality.ExpandID(pnID.(string)))
	}

	f, err := api.CreateFunction(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if zipFile, zipFileExists := d.GetOk("zip_file"); zipFileExists {
		err = functionUpload(ctx, m, api, region, f.ID, zipFile.(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to upload function",
				Detail:   err.Error(),
			})
		}
	}

	if d.Get("deploy").(bool) {
		err = functionDeploy(ctx, api, region, f.ID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to deploy function",
				Detail:   err.Error(),
			})
		}
	}

	if f.ErrorMessage != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Function error",
			Detail:   *f.ErrorMessage,
		})
	}

	if f.RuntimeMessage != "" {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Function runtime warning",
			Detail:        f.RuntimeMessage,
			AttributePath: cty.GetAttrPath("runtime"),
		})
	}

	d.SetId(regional.NewIDString(region, f.ID))

	_, err = waitForFunction(ctx, api, region, f.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return append(diags, ResourceFunctionRead(ctx, d, m)...)
}

func ResourceFunctionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	f, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if f.ErrorMessage != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Function error",
			Detail:   *f.ErrorMessage,
		})
	}

	if f.RuntimeMessage != "" {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Function runtime warning",
			Detail:        f.RuntimeMessage,
			AttributePath: cty.GetAttrPath("runtime"),
		})
	}

	_ = d.Set("description", f.Description)
	_ = d.Set("environment_variables", f.EnvironmentVariables)
	_ = d.Set("handler", f.Handler)
	_ = d.Set("max_scale", int(f.MaxScale))
	_ = d.Set("memory_limit", int(f.MemoryLimit))
	_ = d.Set("cpu_limit", int(f.CPULimit))
	_ = d.Set("min_scale", int(f.MinScale))
	_ = d.Set("name", f.Name)
	_ = d.Set("privacy", f.Privacy.String())
	_ = d.Set("region", f.Region.String())
	_ = d.Set("timeout", f.Timeout.Seconds)
	_ = d.Set("domain_name", f.DomainName)
	_ = d.Set("http_option", f.HTTPOption)
	_ = d.Set("namespace_id", f.NamespaceID)
	_ = d.Set("sandbox", f.Sandbox)
	_ = d.Set("secret_environment_variables", flattenFunctionSecrets(f.SecretEnvironmentVariables))
	_ = d.Set("tags", types.FlattenSliceString(f.Tags))

	if f.PrivateNetworkID != nil {
		_ = d.Set("private_network_id", regional.NewID(region, types.FlattenStringPtr(f.PrivateNetworkID).(string)).String())
	} else {
		_ = d.Set("private_network_id", nil)
	}

	return diags
}

func ResourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	f, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &function.UpdateFunctionRequest{
		Region:     region,
		FunctionID: f.ID,
	}
	updated := false

	if d.HasChange("environment_variables") {
		req.EnvironmentVariables = types.ExpandMapPtrStringString(d.Get("environment_variables"))
		updated = true
	}

	if d.HasChanges("secret_environment_variables") {
		oldEnv, newEnv := d.GetChange("secret_environment_variables")
		req.SecretEnvironmentVariables = filterSecretEnvsToPatch(expandFunctionsSecrets(oldEnv), expandFunctionsSecrets(newEnv))
		updated = true
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		updated = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = types.ExpandUint32Ptr(d.Get("memory_limit"))
		updated = true
	}

	if d.HasChange("handler") {
		req.Handler = types.ExpandStringPtr(d.Get("handler").(string))
		updated = true
	}

	if d.HasChange("min_scale") {
		req.MinScale = types.ExpandUint32Ptr(d.Get("min_scale"))
		updated = true
	}

	if d.HasChange("max_scale") {
		req.MaxScale = types.ExpandUint32Ptr(d.Get("max_scale"))
		updated = true
	}

	if d.HasChange("timeout") {
		req.Timeout = &scw.Duration{Seconds: int64(d.Get("timeout").(int))}
		updated = true
	}

	if d.HasChange("http_option") {
		req.HTTPOption = function.FunctionHTTPOption(d.Get("http_option").(string))
		updated = true
	}

	if d.HasChange("privacy") {
		req.Privacy = function.FunctionPrivacy(d.Get("privacy").(string))
		updated = true
	}

	if d.HasChange("sandbox") {
		req.Sandbox = function.FunctionSandbox(d.Get("sandbox").(string))
		updated = true
	}

	if d.HasChange("runtime") {
		req.Runtime = function.FunctionRuntime(d.Get("runtime").(string))
		updated = true
	}

	if d.HasChanges("private_network_id") {
		req.PrivateNetworkID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("private_network_id")))
		updated = true
	}

	if updated {
		_, err = api.UpdateFunction(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		// Function is not in transit state at this point, api did not update it instantly when processing UpdateFunction
		// We sleep so api has time to change resource to a transit state
		// lintignore:R018
		time.Sleep(defaultFunctionAfterUpdateWait)
	}

	zipHasChanged := d.HasChanges("zip_hash", "zip_file")
	deploy := d.Get("deploy").(bool)

	if zipHasChanged {
		err = functionUpload(ctx, m, api, region, f.ID, d.Get("zip_file").(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to upload function: %w", err))
		}
	}

	// deploy only in some conditions
	shouldDeploy := deploy
	shouldDeploy = shouldDeploy || (zipHasChanged && deploy)
	shouldDeploy = shouldDeploy || d.HasChange("runtime")

	if shouldDeploy {
		_, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return nil
		}

		err = functionDeploy(ctx, api, region, f.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceFunctionRead(ctx, d, m)
}

func ResourceFunctionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return nil
	}

	_, err = api.DeleteFunction(&function.DeleteFunctionRequest{
		FunctionID: id,
		Region:     region,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
