package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewaySecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewaySecretCreate,
		ReadContext:   resourceScalewaySecretRead,
		UpdateContext: resourceScalewaySecretUpdate,
		DeleteContext: resourceScalewaySecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultSecretTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the secret",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] associated to secret",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the secret",
			},
			"version_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of versions for this Secret",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewaySecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	secretCreateRequest := &secret.CreateSecretRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
	}

	rawTag, tagExist := d.GetOk("tags")
	if tagExist {
		secretCreateRequest.Tags = expandStrings(rawTag)
	}

	rawDescription, descriptionExist := d.GetOk("description")
	if descriptionExist {
		secretCreateRequest.Description = expandStringPtr(rawDescription)
	}

	secretResponse, err := api.CreateSecret(secretCreateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, secretResponse.ID))

	return resourceScalewaySecretRead(ctx, d, meta)
}

func resourceScalewaySecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := secretAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	secretResponse, err := api.GetSecret(&secret.GetSecretRequest{
		Region:   region,
		SecretID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if len(secretResponse.Tags) > 0 {
		_ = d.Set("tags", flattenSliceString(secretResponse.Tags))
	}

	_ = d.Set("name", secretResponse.Name)
	_ = d.Set("description", flattenStringPtr(secretResponse.Description))
	_ = d.Set("created_at", flattenTime(secretResponse.CreatedAt))
	_ = d.Set("updated_at", flattenTime(secretResponse.UpdatedAt))
	_ = d.Set("status", secretResponse.Status.String())
	_ = d.Set("version_count", int(secretResponse.VersionCount))
	_ = d.Set("region", string(region))
	_ = d.Set("project_id", secretResponse.ProjectID)

	return nil
}

func resourceScalewaySecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := secretAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &secret.UpdateSecretRequest{
		Region:   region,
		SecretID: id,
	}

	hasChanged := false

	if d.HasChange("description") {
		updateRequest.Description = expandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("name") {
		updateRequest.Name = expandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = expandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateSecret(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewaySecretRead(ctx, d, meta)
}

func resourceScalewaySecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := secretAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteSecret(&secret.DeleteSecretRequest{
		Region:   region,
		SecretID: id,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
