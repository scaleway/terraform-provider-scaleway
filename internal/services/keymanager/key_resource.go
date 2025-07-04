package keymanager

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceKeyManagerKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyManagerKeyCreate,
		ReadContext:   resourceKeyManagerKeyRead,
		UpdateContext: resourceKeyManagerKeyUpdate,
		DeleteContext: resourceKeyManagerKeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
			"usage": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"symmetric_encryption", "asymmetric_encryption", "asymmetric_signing",
				}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rotation_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rotation_period":  {Type: schema.TypeString, Optional: true},
						"next_rotation_at": {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"unprotected": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"origin": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"scaleway_kms", "external",
				}, false),
			},
			// Computed fields
			"id":             {Type: schema.TypeString, Computed: true},
			"state":          {Type: schema.TypeString, Computed: true},
			"created_at":     {Type: schema.TypeString, Computed: true},
			"updated_at":     {Type: schema.TypeString, Computed: true},
			"rotation_count": {Type: schema.TypeInt, Computed: true},
			"protected":      {Type: schema.TypeBool, Computed: true},
			"locked":         {Type: schema.TypeBool, Computed: true},
			"rotated_at":     {Type: schema.TypeString, Computed: true},
			"origin_read":    {Type: schema.TypeString, Computed: true},
			"region_read":    {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceKeyManagerKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := key_manager.NewAPI(meta.ExtractScwClient(m))

	region := scw.Region(d.Get("region").(string))
	projectID := d.Get("project_id").(string)
	name := d.Get("name").(string)
	usage := d.Get("usage").(string)
	description := d.Get("description").(string)
	unprotected := d.Get("unprotected").(bool)
	origin := d.Get("origin").(string)
	tags := ExpandStringList(d.Get("tags"))

	var usageBlock *key_manager.KeyUsage

	switch usage {
	case "symmetric_encryption":
		alg := key_manager.KeyAlgorithmSymmetricEncryptionAes256Gcm
		usageBlock = &key_manager.KeyUsage{SymmetricEncryption: &alg}
	case "asymmetric_encryption":
		alg := key_manager.KeyAlgorithmAsymmetricEncryptionRsaOaep3072Sha256
		usageBlock = &key_manager.KeyUsage{AsymmetricEncryption: &alg}
	case "asymmetric_signing":
		alg := key_manager.KeyAlgorithmAsymmetricSigningEcP256Sha256
		usageBlock = &key_manager.KeyUsage{AsymmetricSigning: &alg}
	default:
		return diag.Errorf("invalid usage: %s", usage)
	}

	var rotationPolicy *key_manager.KeyRotationPolicy

	if v, ok := d.GetOk("rotation_policy"); ok && len(v.([]interface{})) > 0 {
		m := v.([]interface{})[0].(map[string]interface{})

		if period, ok := m["rotation_period"].(string); ok && period != "" {
			dur, err := time.ParseDuration(period)
			if err != nil {
				return diag.Errorf("invalid rotation_period: %v", err)
			}

			sdur := scw.NewDurationFromTimeDuration(dur)
			rotationPolicy = &key_manager.KeyRotationPolicy{
				RotationPeriod: sdur,
			}
		}
	}

	createReq := &key_manager.CreateKeyRequest{
		Region:         region,
		ProjectID:      projectID,
		Name:           &name,
		Usage:          usageBlock,
		Description:    &description,
		Tags:           tags,
		RotationPolicy: rotationPolicy,
		Unprotected:    unprotected,
		Origin:         key_manager.KeyOrigin(origin),
	}

	key, err := client.CreateKey(createReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", key.Region, key.ID))

	return resourceKeyManagerKeyRead(ctx, d, m)
}

func resourceKeyManagerKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	_ = d.Set("usage", UsageToString(key.Usage))
	_ = d.Set("description", key.Description)
	_ = d.Set("tags", key.Tags)
	_ = d.Set("rotation_count", int(key.RotationCount))
	_ = d.Set("created_at", TimeToRFC3339(key.CreatedAt))
	_ = d.Set("updated_at", TimeToRFC3339(key.UpdatedAt))
	_ = d.Set("protected", key.Protected)
	_ = d.Set("locked", key.Locked)
	_ = d.Set("rotated_at", TimeToRFC3339(key.RotatedAt))
	_ = d.Set("origin_read", key.Origin.String())
	_ = d.Set("region_read", key.Region.String())

	if key.RotationPolicy != nil {
		var periodStr string

		if key.RotationPolicy.RotationPeriod != nil {
			periodStr = key.RotationPolicy.RotationPeriod.ToTimeDuration().String()
		}

		_ = d.Set("rotation_policy", []map[string]interface{}{
			{
				"rotation_period":  periodStr,
				"next_rotation_at": TimeToRFC3339(key.RotationPolicy.NextRotationAt),
			},
		})
	}

	return nil
}

func resourceKeyManagerKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		tags := ExpandStringList(d.Get("tags"))
		updateReq.Tags = &tags
	}

	if d.HasChange("rotation_policy") {
		if v, ok := d.GetOk("rotation_policy"); ok && len(v.([]interface{})) > 0 {
			m := v.([]interface{})[0].(map[string]interface{})

			if period, ok := m["rotation_period"].(string); ok && period != "" {
				dur, err := time.ParseDuration(period)
				if err != nil {
					return diag.Errorf("invalid rotation_period: %v", err)
				}

				sdur := scw.NewDurationFromTimeDuration(dur)
				updateReq.RotationPolicy = &key_manager.KeyRotationPolicy{
					RotationPeriod: sdur,
				}
			}
		}
	}

	_, err = client.UpdateKey(updateReq)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeyManagerKeyRead(ctx, d, m)
}

func resourceKeyManagerKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
