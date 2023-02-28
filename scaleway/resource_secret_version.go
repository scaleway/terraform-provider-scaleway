package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewaySecretVersion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewaySecretVersionCreate,
		ReadContext:   resourceScalewaySecretVersionRead,
		UpdateContext: resourceScalewaySecretVersionUpdate,
		DeleteContext: resourceScalewaySecretVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultSecretTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The secret ID associated with this version",
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The data payload of your secret version",
				Sensitive:   true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the secret version",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the secret version",
			},
			"revision": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The revision of secret version",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret version's creation (RFC 3339 format)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret version's creation (RFC 3339 format)",
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewaySecretVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID := expandID(d.Get("secret_id").(string))
	secretCreateVersionRequest := &secret.CreateSecretVersionRequest{
		Region:      region,
		SecretID:    secretID,
		Data:        []byte(d.Get("data").(string)),
		Description: expandStringPtr(d.Get("description")),
	}

	rawDescription, descriptionExist := d.GetOk("description")
	if descriptionExist {
		secretCreateVersionRequest.Description = expandStringPtr(rawDescription)
	}

	secretResponse, err := api.CreateSecretVersion(secretCreateVersionRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("data", d.Get("data").(string))

	d.SetId(newRegionalIDString(region, fmt.Sprintf("%s/%d", secretResponse.SecretID, secretResponse.Revision)))

	return resourceScalewaySecretVersionRead(ctx, d, meta)
}

func resourceScalewaySecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, revision, err := secretVersionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	secretResponse, err := api.GetSecretVersion(&secret.GetSecretVersionRequest{
		Region:   region,
		SecretID: id,
		Revision: revision,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("secret_id", newRegionalIDString(region, id))
	_ = d.Set("description", flattenStringPtr(secretResponse.Description))
	_ = d.Set("created_at", flattenTime(secretResponse.CreatedAt))
	_ = d.Set("updated_at", flattenTime(secretResponse.UpdatedAt))
	_ = d.Set("status", secretResponse.Status.String())
	_ = d.Set("revision", int(secretResponse.Revision))
	_ = d.Set("region", string(region))

	return nil
}

func resourceScalewaySecretVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, revision, err := secretVersionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &secret.UpdateSecretVersionRequest{
		Region:   region,
		SecretID: id,
		Revision: revision,
	}

	hasChanged := false

	if d.HasChange("description") {
		updateRequest.Description = expandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateSecretVersion(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewaySecretVersionRead(ctx, d, meta)
}

func resourceScalewaySecretVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, revision, err := secretVersionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DestroySecretVersion(&secret.DestroySecretVersionRequest{
		Region:   region,
		SecretID: id,
		Revision: revision,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
