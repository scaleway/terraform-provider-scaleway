package scaleway

import (
	"context"
	"fmt"

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
				Description:  "The name of the function namespace",
				ValidateFunc: validation.StringLenBetween(1, 20),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the function namespace",
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
			"privacy": {
				Type:        schema.TypeString,
				Description: "Privacy of the function namespace. Can be either `private` or `public`",
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					function.FunctionPrivacyPublic.String(),
					function.FunctionPrivacyPrivate.String(),
				}, false),
			},
			"runtime": {
				Type:        schema.TypeString,
				Description: "Runtime of the function namespace",
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					function.FunctionRuntimeGolang.String(),
					function.FunctionRuntimeNode8.String(),
					function.FunctionRuntimeNode10.String(),
					function.FunctionRuntimeNode14.String(),
					function.FunctionRuntimePython.String(),
					function.FunctionRuntimePython3.String(),
					function.FunctionRuntime("go118").String(),
				}, false),
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
				Description: "Define if the function should be deployed on upload, terraform will wait for function to be deployed",
			},

			/*
				"status": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The state of the function, possible values in api doc https://developers.scaleway.com/en/products/functions/api/#status-1e9767",
				},
				"cpu_limit": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "CPU limit in mCPU for your function, defaults to 70mCPU",
					Default:     70,
				},
				"secret_environment_variables": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						Sensitive:    true,
						ValidateFunc: validation.StringLenBetween(0, 1000),
					},
					ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
				},
			*/
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

	f, err := api.WaitForFunction(&function.WaitForFunctionRequest{
		FunctionID: id,
		Region:     region,
	}, scw.WithContext(ctx))
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

	f, err := api.WaitForFunction(&function.WaitForFunctionRequest{
		FunctionID: id,
		Region:     region,
	}, scw.WithContext(ctx))
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
	update := false

	if d.HasChange("environment_variables") {
		req.EnvironmentVariables = expandMapStringStringPtr(d.Get("environment_variables"))
		update = true
	}

	if d.HasChange("description") {
		req.Description = expandStringPtr(d.Get("description"))
		update = true
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = expandUint32Ptr(d.Get("memory_limit"))
		update = true
	}

	if d.HasChange("handler") {
		req.Handler = expandStringPtr(d.Get("handler").(string))
		update = true
	}

	if d.HasChange("min_scale") {
		req.MinScale = expandUint32Ptr(d.Get("min_scale"))
		update = true
	}

	if d.HasChange("max_scale") {
		req.MinScale = expandUint32Ptr(d.Get("max_scale"))
		update = true
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = expandUint32Ptr(d.Get("memory_limit"))
		update = true
	}

	if d.HasChange("timeout") {
		req.Timeout = &scw.Duration{Seconds: d.Get("timeout").(int64)}
		update = true
	}

	if update {
		_, err = api.UpdateFunction(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("zip_hash") || d.HasChange("zip_file") {
		// Function is not in transit state at this point, api did not update it instantly when processing UpdateFunction
		_, err = api.WaitForFunction(&function.WaitForFunctionRequest{
			FunctionID: id,
			Region:     region,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil
		}
		err = functionUpload(ctx, meta, api, region, f.ID, d.Get("zip_file").(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to upload function: %w", err))
		}
		if d.Get("deploy").(bool) {
			err = functionDeploy(ctx, api, region, f.ID)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceScalewayFunctionRead(ctx, d, meta)
}

func resourceScalewayFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.WaitForFunction(&function.WaitForFunctionRequest{
		FunctionID: id,
		Region:     region,
	}, scw.WithContext(ctx))
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
