package scaleway

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
)

func resourceScalewayFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFunctionCreate,
		ReadContext:   resourceScalewayFunctionRead,
		UpdateContext: resourceScalewayFunctionUpdate,
		DeleteContext: resourceScalewayFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultFunctionTimeout),
			Read:    schema.DefaultTimeout(defaultFunctionTimeout),
			Update:  schema.DefaultTimeout(defaultFunctionTimeout),
			Delete:  schema.DefaultTimeout(defaultFunctionTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"namespace_id": {
				Type:        schema.TypeString,
				Description: "The namespace ID associated with this function",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				Description:  "The name of the function",
				ValidateFunc: validation.StringLenBetween(1, 20),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the function",
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
			"privacy": {
				Type:        schema.TypeString,
				Description: "Privacy of the function. Can be either `private` or `public`",
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					function.FunctionPrivacyPublic.String(),
					function.FunctionPrivacyPrivate.String(),
				}, false),
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
				Description: "Memory limit in MB for your function, defaults to 128MB",
				Optional:    true,
				Default:     128,
			},
			"handler": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Handler of the function. Depends on the runtime https://developers.scaleway.com/en/products/functions/api/#create-a-function",
			},
			"timeout": {
				Type:        schema.TypeString,
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
				Description:  "The hash of your source zip file, changing it will re-apply function",
			},
			"deploy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Define if the function should be deployed, terraform will wait for function to be deployed",
			},
			"cpu_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU limit in mCPU for your function",
			},

			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, namespace, err := parseRegionalID(d.Get("namespace_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &function.CreateFunctionRequest{
		Description:          expandStringPtr(d.Get("description").(string)),
		EnvironmentVariables: expandMapStringStringPtr(d.Get("environment_variables")),
		Handler:              expandStringPtr(d.Get("handler").(string)),
		MaxScale:             expandUint32Ptr(d.Get("max_scale")),
		MemoryLimit:          expandUint32Ptr(d.Get("memory_limit")),
		MinScale:             expandUint32Ptr(d.Get("min_scale")),
		Name:                 expandOrGenerateString(d.Get("name").(string), "func"),
		NamespaceID:          namespace,
		Privacy:              function.FunctionPrivacy(d.Get("privacy").(string)),
		Region:               region,
		Runtime:              function.FunctionRuntime(d.Get("runtime").(string)),
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		req.Timeout = &scw.Duration{Seconds: timeout.(int64)}
	}

	f, err := api.CreateFunction(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if zipFile, zipFileExists := d.GetOk("zip_file"); zipFileExists {
		err = functionUpload(ctx, meta, api, region, f.ID, zipFile.(string))
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

	d.SetId(newRegionalIDString(region, f.ID))

	return append(diags, resourceScalewayFunctionRead(ctx, d, meta)...)
}

func resourceScalewayFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	f, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
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
	_ = d.Set("timeout", flattenDuration(f.Timeout.ToTimeDuration()))

	return diags
}

func resourceScalewayFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	f, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
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
		req.EnvironmentVariables = expandMapStringStringPtr(d.Get("environment_variables"))
		updated = true
	}

	if d.HasChange("description") {
		req.Description = expandStringPtr(d.Get("description"))
		updated = true
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = expandUint32Ptr(d.Get("memory_limit"))
		updated = true
	}

	if d.HasChange("handler") {
		req.Handler = expandStringPtr(d.Get("handler").(string))
		updated = true
	}

	if d.HasChange("min_scale") {
		req.MinScale = expandUint32Ptr(d.Get("min_scale"))
		updated = true
	}

	if d.HasChange("max_scale") {
		req.MaxScale = expandUint32Ptr(d.Get("max_scale"))
		updated = true
	}

	if d.HasChange("timeout") {
		req.Timeout = &scw.Duration{Seconds: d.Get("timeout").(int64)}
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
	shouldDeploy := d.Get("deploy").(bool)

	if zipHasChanged {
		err = functionUpload(ctx, meta, api, region, f.ID, d.Get("zip_file").(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to upload function: %w", err))
		}
	}

	if d.HasChange("deploy") && shouldDeploy || zipHasChanged && shouldDeploy {
		_, err := waitForFunction(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return nil
		}
		err = functionDeploy(ctx, api, region, f.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayFunctionRead(ctx, d, meta)
}

func resourceScalewayFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
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

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
