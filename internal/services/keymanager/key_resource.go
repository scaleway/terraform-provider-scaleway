package keymanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceKeyManagerKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyManagerKeyCreate,
		ReadContext:   resourceKeyManagerKeyRead,
		UpdateContext: resourceKeyManagerKeyUpdate,
		DeleteContext: resourceKeyManagerKeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the key.",
			},
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
			"usage": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"symmetric_encryption", "asymmetric_encryption", "asymmetric_signing",
				}, false),
				Deprecated:  "Use usage_symmetric_encryption, usage_asymmetric_encryption, or usage_asymmetric_signing instead",
				Description: "DEPRECATED: Use usage_symmetric_encryption, usage_asymmetric_encryption, or usage_asymmetric_signing instead. Key usage. Possible values: symmetric_encryption, asymmetric_encryption, asymmetric_signing.",
				ExactlyOneOf: []string{
					"usage",
					"usage_symmetric_encryption",
					"usage_asymmetric_encryption",
					"usage_asymmetric_signing",
				},
			},
			"usage_symmetric_encryption": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"aes_256_gcm",
				}, false),
				Description: "Algorithm for symmetric encryption. Possible values: aes_256_gcm",
				ExactlyOneOf: []string{
					"usage",
					"usage_symmetric_encryption",
					"usage_asymmetric_encryption",
					"usage_asymmetric_signing",
				},
			},
			"usage_asymmetric_encryption": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"rsa_oaep_2048_sha256",
					"rsa_oaep_3072_sha256",
					"rsa_oaep_4096_sha256",
				}, false),
				Description: "Algorithm for asymmetric encryption. Possible values: rsa_oaep_2048_sha256, rsa_oaep_3072_sha256, rsa_oaep_4096_sha256",
				ExactlyOneOf: []string{
					"usage",
					"usage_symmetric_encryption",
					"usage_asymmetric_encryption",
					"usage_asymmetric_signing",
				},
			},
			"usage_asymmetric_signing": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ec_p256_sha256",
					"rsa_pss_2048_sha256",
					"rsa_pkcs1_2048_sha256",
				}, false),
				Description: "Algorithm for asymmetric signing. Possible values: ec_p256_sha256, rsa_pss_2048_sha256, rsa_pkcs1_2048_sha256",
				ExactlyOneOf: []string{
					"usage",
					"usage_symmetric_encryption",
					"usage_asymmetric_encryption",
					"usage_asymmetric_signing",
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the key.",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of the key's tags.",
			},
			"rotation_policy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Key rotation policy.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rotation_period":  {Type: schema.TypeString, Required: true, DiffSuppressFunc: dsf.Duration, Description: "Time interval between two key rotations. The minimum duration is 24 hours and the maximum duration is 1 year (876000 hours)."},
						"next_rotation_at": {Type: schema.TypeString, Optional: true, Description: "Timestamp indicating the next scheduled rotation."},
					},
				},
			},
			"unprotected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, the key is not protected against deletion.",
			},
			"origin": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"scaleway_kms", "external",
				}, false),
				Description: "Origin of the key material. Possible values: scaleway_kms (Key Manager generates the key material), external (key material comes from an external source).",
			},
			// Computed fields
			"id":             {Type: schema.TypeString, Computed: true, Description: "ID of the key."},
			"state":          {Type: schema.TypeString, Computed: true, Description: "State of the key. See the Key.State enum for possible values."},
			"created_at":     {Type: schema.TypeString, Computed: true, Description: "Key creation date."},
			"updated_at":     {Type: schema.TypeString, Computed: true, Description: "Key last modification date."},
			"rotation_count": {Type: schema.TypeInt, Computed: true, Description: "The rotation count tracks the number of times the key has been rotated."},
			"protected":      {Type: schema.TypeBool, Computed: true, Description: "Returns true if key protection is applied to the key."},
			"locked":         {Type: schema.TypeBool, Computed: true, Description: "Returns true if the key is locked."},
			"rotated_at":     {Type: schema.TypeString, Computed: true, Description: "Key last rotation date."},
		},
	}
}

func resourceKeyManagerKeyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newKeyManagerAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &key_manager.CreateKeyRequest{
		Region:      region,
		ProjectID:   d.Get("project_id").(string),
		Name:        types.ExpandStringPtr(d.Get("name")),
		Description: types.ExpandStringPtr(d.Get("description")),
		Unprotected: d.Get("unprotected").(bool),
	}

	if v, ok := d.GetOk("tags"); ok {
		createReq.Tags = types.ExpandStrings(v)
	}

	if v, ok := d.GetOk("rotation_policy"); ok {
		rp, err := ExpandKeyRotationPolicy(v)
		if err != nil {
			return diag.Errorf("invalid rotation_period: %v", err)
		}

		createReq.RotationPolicy = rp
	}

	if v, ok := d.GetOk("origin"); ok {
		createReq.Origin = key_manager.KeyOrigin(v.(string))
	}

	createReq.Usage = ExpandKeyUsageFromFields(d)

	key, err := api.CreateKey(createReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(key.Region, key.ID))

	return resourceKeyManagerKeyRead(ctx, d, m)
}

func resourceKeyManagerKeyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client, region, keyID, err := NewKeyManagerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := client.GetKey(&key_manager.GetKeyRequest{
		Region: region,
		KeyID:  keyID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", key.Name)
	_ = d.Set("project_id", key.ProjectID)
	_ = d.Set("region", key.Region.String())

	usageType := UsageToString(key.Usage)
	algorithm := AlgorithmFromKeyUsage(key.Usage)

	_ = d.Set("usage", usageType)

	_, usesLegacy := d.GetOk("usage")

	if !usesLegacy {
		switch usageType {
		case "symmetric_encryption":
			_ = d.Set("usage_symmetric_encryption", algorithm)
		case "asymmetric_encryption":
			_ = d.Set("usage_asymmetric_encryption", algorithm)
		case "asymmetric_signing":
			_ = d.Set("usage_asymmetric_signing", algorithm)
		}
	}

	_ = d.Set("description", key.Description)
	_ = d.Set("tags", key.Tags)
	_ = d.Set("rotation_count", int(key.RotationCount))
	_ = d.Set("created_at", types.FlattenTime(key.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(key.UpdatedAt))
	_ = d.Set("protected", key.Protected)
	_ = d.Set("locked", key.Locked)
	_ = d.Set("rotated_at", types.FlattenTime(key.RotatedAt))
	_ = d.Set("rotation_policy", FlattenKeyRotationPolicy(key.RotationPolicy))

	return nil
}

func resourceKeyManagerKeyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client, region, keyID, err := NewKeyManagerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateReq := &key_manager.UpdateKeyRequest{
		Region: region,
		KeyID:  keyID,
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		updateReq.Name = &name
	}

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		updateReq.Description = &desc
	}

	if d.HasChange("tags") {
		tags := types.ExpandStrings(d.Get("tags"))
		updateReq.Tags = &tags
	}

	if d.HasChange("rotation_policy") {
		if v, ok := d.GetOk("rotation_policy"); ok {
			rp, err := ExpandKeyRotationPolicy(v)
			if err != nil {
				return diag.Errorf("invalid rotation_period: %v", err)
			}

			updateReq.RotationPolicy = rp
		}
	}

	_, err = client.UpdateKey(updateReq)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeyManagerKeyRead(ctx, d, m)
}

func resourceKeyManagerKeyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client, region, keyID, err := NewKeyManagerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteKey(&key_manager.DeleteKeyRequest{
		Region: region,
		KeyID:  keyID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
