package inference

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceCustomModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCustomModelCreate,
		ReadContext:   ResourceCustomModelRead,
		DeleteContext: ResourceCustomModelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultCustomModelTimeout),
			Create:  schema.DefaultTimeout(defaultCustomModelTimeout),
			Update:  schema.DefaultTimeout(defaultCustomModelTimeout),
			Delete:  schema.DefaultTimeout(defaultCustomModelTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the model",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The URL of the model",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "The secret to pull a model",
			},
			"tags": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "The tags associated with the deployment",
			},
			"project_id": account.ProjectIDSchema(),
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the model",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the model",
			},
			"error_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Displays information if your model is in error state",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the model",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the model",
			},
			"has_eula": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether the model has an end user license agreement",
			},
			"nodes_support": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Supported node types with quantization options and context lengths.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_type_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Supported node type.",
						},
						"quantization": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Supported quantization options.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"quantization_bits": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Number of bits used for quantization.",
									},
									"allowed": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether this quantization is allowed for the model.",
									},
									"max_context_size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum inference context size for this quantization and node type.",
									},
								},
							},
						},
					},
				},
			},
			"parameter_size_bits": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Size, in bits, of the model parameters",
			},
			"size_bits": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total size, in bytes, of the model files",
			},
			"region": regional.Schema(),
		},
	}
}

func ResourceCustomModelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	modelSource := &inference.ModelSource{
		URL: d.Get("url").(string),
	}
	if secret, ok := d.GetOk("secret"); ok {
		secretStr := secret.(string)
		modelSource.Secret = &secretStr
	}

	reqCreateModel := &inference.CreateModelRequest{
		Region:    region,
		Name:      d.Get("name").(string),
		ProjectID: d.Get("project_id").(string),
		Source:    modelSource,
	}

	model, err := api.CreateModel(reqCreateModel)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, model.ID))

	model, err = waitForModel(ctx, api, region, model.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if model.Status == inference.ModelStatusError {
		errMsg := *model.ErrorMessage
		return diag.FromErr(fmt.Errorf("model '%s' is in status '%s'", model.ID, errMsg))
	}

	return ResourceCustomModelRead(ctx, d, m)
}

func ResourceCustomModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	model, err := waitForModel(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}
	
	_ = d.Set("parameter_size_bits", model.ParameterSizeBits)
	_ = d.Set("size_bits", model.SizeBytes)
	_ = d.Set("name", model.Name)
	_ = d.Set("status", model.Status)
	_ = d.Set("description", model.Description)
	_ = d.Set("tags", types.ExpandUpdatedStringsPtr(model.Tags))
	_ = d.Set("created_at", types.FlattenTime(model.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(model.UpdatedAt))
	_ = d.Set("has_eula", model.HasEula)
	_ = d.Set("nodes_support", flattenNodeSupport(model.NodesSupport))
	_ = d.Set("error_message", model.ErrorMessage)

	return nil
}

func ResourceCustomModelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForModel(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteModel(&inference.DeleteModelRequest{
		Region:  region,
		ModelID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
