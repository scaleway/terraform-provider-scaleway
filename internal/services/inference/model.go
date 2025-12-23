package inference

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceModelCreate,
		ReadContext:   ResourceModelRead,
		DeleteContext: ResourceModelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultModelTimeout),
			Create:  schema.DefaultTimeout(defaultModelTimeout),
			Delete:  schema.DefaultTimeout(defaultModelTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    modelSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func modelSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the model",
		},
		"url": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: "The HTTPS URL to the model archive or repository. Typically, this is a Hugging Face repository URL (e.g., " +
				"`https://huggingface.co/your-org/your-model`). The URL must be publicly accessible or require a valid secret for authentication.",
		},
		"secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			ForceNew:    true,
			Description: "A token or credential used to authenticate when pulling the model from a private or gated source. For example, a Hugging Face access token with read permissions.",
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
		"size_bytes": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Total size, in bytes, of the model files",
		},
		"region": regional.Schema(),
	}
}

func ResourceModelCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	modelSource := &inference.ModelSource{
		URL: d.Get("url").(string),
	}

	if secret, ok := d.GetOk("secret"); ok {
		modelSource.Secret = types.ExpandStringPtr(secret)
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

	return ResourceModelRead(ctx, d, m)
}

func ResourceModelRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	_ = d.Set("parameter_size_bits", int32(model.ParameterSizeBits))
	_ = d.Set("size_bytes", int64(model.SizeBytes))
	_ = d.Set("name", model.Name)
	_ = d.Set("status", model.Status.String())
	_ = d.Set("description", model.Description)
	_ = d.Set("tags", model.Tags)
	_ = d.Set("created_at", types.FlattenTime(model.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(model.UpdatedAt))
	_ = d.Set("has_eula", model.HasEula)
	_ = d.Set("nodes_support", flattenNodeSupport(model.NodesSupport))

	return nil
}

func ResourceModelDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
