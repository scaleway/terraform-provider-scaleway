package scaleway

import (
	"context"

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
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The name of the function namespace",
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
			"http_option": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTPOption: configure how HTTP and HTTPS requests are handled (redirected or enabled)",
				ValidateFunc: validation.StringInSlice([]string{
					"redirected", // Responds to HTTP request with a 302 redirect to ask the clients to use HTTPS.
					"enabled",    // Serve both HTTP and HTTPS traffic.
				}, false),
			},
			"timeout": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Holds the max duration (in seconds) the function is allowed for responding to a request",
				Optional:    true,
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
		HTTPOption:           expandStringPtr(d.Get("http_option").(string)),
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

	d.SetId(newRegionalIDString(region, f.ID))

	return resourceScalewayFunctionRead(ctx, d, meta)
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

	_ = d.Set("description", f.Description)
	_ = d.Set("environment_variables", f.EnvironmentVariables)
	_ = d.Set("handler", f.Handler)
	_ = d.Set("http_option", f.HTTPOption)
	_ = d.Set("max_scale", int(f.MaxScale))
	_ = d.Set("memory_limit", int(f.MemoryLimit))
	_ = d.Set("min_scale", int(f.MinScale))
	_ = d.Set("privacy", f.Privacy.String())
	_ = d.Set("region", f.Region.String())
	_ = d.Set("timeout", flattenDuration(f.Timeout.ToTimeDuration()))

	return nil
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

	if d.HasChange("environment_variables") {
		req.EnvironmentVariables = expandMapStringStringPtr(d.Get("environment_variables"))
	}

	if d.HasChange("description") {
		req.Description = expandStringPtr(d.Get("description"))
	}

	if d.HasChange("http_option") {
		req.HTTPOption = expandStringPtr(d.Get("http_option"))
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = expandUint32Ptr(d.Get("memory_limit"))
	}

	if d.HasChange("handler") {
		req.Handler = expandStringPtr(d.Get("handler").(string))
	}

	if d.HasChange("min_scale") {
		req.MinScale = expandUint32Ptr(d.Get("min_scale"))
	}

	if d.HasChange("max_scale") {
		req.MinScale = expandUint32Ptr(d.Get("max_scale"))
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = expandUint32Ptr(d.Get("memory_limit"))
	}

	if d.HasChange("timeout") {
		req.Timeout = &scw.Duration{Seconds: d.Get("timeout").(int64)}
	}

	_, err = api.UpdateFunction(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
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
