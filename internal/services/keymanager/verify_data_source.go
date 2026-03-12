package keymanager

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/verify_data_source.md
var verifyDataSourceDescription string

func DataSourceVerify() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVerifyRead,
		SchemaFunc:  dataSourceVerifySchema,
		Description: verifyDataSourceDescription,
	}
}

func dataSourceVerifySchema() map[string]*schema.Schema {
	dsSchema := map[string]*schema.Schema{
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration.",
		},
		"key_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "ID of the key to use for signature verification. Can be a plain UUID or a regional ID.",
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		},
		"digest": {
			Type:     schema.TypeString,
			Required: true,
			// Digest could contain non-UTF-8 bytes, so we ask for base64 encoding.
			Description: "Digest of the original signed message. Must be generated using the same algorithm specified in the key’s configuration, and encoded as a base64 string.",
		},
		"signature": {
			Type:     schema.TypeString,
			Required: true,
			// Signature could contain non-UTF-8 bytes, so we ask for base64 encoding.
			Description: "The message signature to verify, encoded as a base64 string.",
		},
		"valid": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Defines whether the signature is valid. Returns `true` if the signature is valid for the digest and key, and `false` otherwise.",
		},
	}

	return dsSchema
}

func dataSourceVerifyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	keyID := d.Get("key_id")

	var region scw.Region

	var err error

	if reg, ok := d.GetOk("region"); ok {
		region = scw.Region(reg.(string))
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(keyID.(string)); parseErr == nil {
			region = derivedRegion
			keyID = id
		} else {
			region, err = meta.ExtractRegion(d, m)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	regionalID := datasource.NewRegionalID(keyID, region)
	d.SetId(regionalID)

	err = d.Set("key_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	client, region, keyID, err := NewKeyManagerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	digestBytes, err := base64.StdEncoding.DecodeString(d.Get("digest").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(d.Get("signature").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to decode signature from base64: %w", err))
	}

	resp, err := client.Verify(&key_manager.VerifyRequest{
		Region:    region,
		KeyID:     keyID.(string),
		Digest:    digestBytes,
		Signature: signatureBytes,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("valid", resp.Valid)

	return nil
}
