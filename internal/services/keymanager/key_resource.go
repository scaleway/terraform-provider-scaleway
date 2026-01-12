package keymanager

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceKeyManagerKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyManagerKeyCreate,
		ReadContext:   resourceKeyManagerKeyRead,
		UpdateContext: resourceKeyManagerKeyUpdate,
		DeleteContext: resourceKeyManagerKeyDelete,
		CustomizeDiff: customdiff.All(
			validateUsageAlgorithmCombination(),
		),
		SchemaFunc: keySchema,
	}
}

func keySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the key.",
		},
		"project_id": account.ProjectIDSchema(),
		"region":     regional.Schema(),
		"usage": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"symmetric_encryption", "asymmetric_encryption", "asymmetric_signing",
			}, false),
			Description: "Key usage type. Possible values: symmetric_encryption, asymmetric_encryption, asymmetric_signing.",
		},
		"algorithm": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Algorithm to use for the key. The valid algorithms depend on the usage type.",
			ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
				symAlgos := key_manager.KeyAlgorithmSymmetricEncryption("").Values()
				asymEncAlgos := key_manager.KeyAlgorithmAsymmetricEncryption("").Values()
				asymSignAlgos := key_manager.KeyAlgorithmAsymmetricSigning("").Values()

				allKnownAlgos := make([]string, 0, len(symAlgos)+len(asymEncAlgos)+len(asymSignAlgos))

				for _, algo := range symAlgos {
					allKnownAlgos = append(allKnownAlgos, string(algo))
				}

				for _, algo := range asymEncAlgos {
					allKnownAlgos = append(allKnownAlgos, string(algo))
				}

				for _, algo := range asymSignAlgos {
					allKnownAlgos = append(allKnownAlgos, string(algo))
				}

				return verify.ValidateStringInSliceWithWarning(allKnownAlgos, "algorithm")(i, p)
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

	usage := d.Get("usage").(string)
	algorithm := d.Get("algorithm").(string)

	keyUsage, err := expandUsageAlgorithm(usage, algorithm)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq.Usage = keyUsage

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
	_ = d.Set("algorithm", algorithm)

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

func validateUsageAlgorithmCombination() schema.CustomizeDiffFunc {
	return func(ctx context.Context, diff *schema.ResourceDiff, _ any) error {
		return nil
	}
}

func expandUsageAlgorithm(usage, algorithm string) (*key_manager.KeyUsage, error) {
	switch usage {
	case usageSymmetricEncryption:
		typedAlgo := key_manager.KeyAlgorithmSymmetricEncryption(algorithm)

		return &key_manager.KeyUsage{SymmetricEncryption: &typedAlgo}, nil

	case usageAsymmetricEncryption:
		typedAlgo := key_manager.KeyAlgorithmAsymmetricEncryption(algorithm)

		return &key_manager.KeyUsage{AsymmetricEncryption: &typedAlgo}, nil

	case usageAsymmetricSigning:
		typedAlgo := key_manager.KeyAlgorithmAsymmetricSigning(algorithm)

		return &key_manager.KeyUsage{AsymmetricSigning: &typedAlgo}, nil

	default:
		return nil, fmt.Errorf("unknown usage type: %s", usage)
	}
}
