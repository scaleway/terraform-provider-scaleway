package keymanager

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func UsageToString(u *key_manager.KeyUsage) string {
	if u == nil {
		return ""
	}

	if u.SymmetricEncryption != nil {
		return "symmetric_encryption"
	}

	if u.AsymmetricEncryption != nil {
		return "asymmetric_encryption"
	}

	if u.AsymmetricSigning != nil {
		return "asymmetric_signing"
	}

	return ""
}

// ExtractRegionAndKeyID parses an ID of the form "region/key_id" and returns the region and key ID.
func ExtractRegionAndKeyID(id string) (scw.Region, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected ID format (%s), expected region/key_id", id)
	}

	return scw.Region(parts[0]), parts[1], nil
}

func newKeyManagerAPI(d *schema.ResourceData, m any) (*key_manager.API, scw.Region, error) {
	api := key_manager.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewKeyManagerAPIWithRegionAndID returns a Key Manager API client, region, and key ID from meta and a composite ID.
func NewKeyManagerAPIWithRegionAndID(m any, id string) (*key_manager.API, scw.Region, string, error) {
	region, keyID, err := ExtractRegionAndKeyID(id)
	if err != nil {
		return nil, "", "", err
	}

	client := key_manager.NewAPI(meta.ExtractScwClient(m))

	return client, region, keyID, nil
}

// ExpandKeyUsage converts a usage string to a *key_manager.KeyUsage struct for API requests.
func ExpandKeyUsage(usage string) *key_manager.KeyUsage {
	// You can extend this switch if you want to support more algorithms in the future
	switch usage {
	case "symmetric_encryption":
		alg := key_manager.KeyAlgorithmSymmetricEncryptionAes256Gcm

		return &key_manager.KeyUsage{SymmetricEncryption: &alg}
	case "asymmetric_encryption":
		alg := key_manager.KeyAlgorithmAsymmetricEncryptionRsaOaep3072Sha256

		return &key_manager.KeyUsage{AsymmetricEncryption: &alg}
	case "asymmetric_signing":
		alg := key_manager.KeyAlgorithmAsymmetricSigningEcP256Sha256

		return &key_manager.KeyUsage{AsymmetricSigning: &alg}
	default:
		return nil
	}
}

// ExpandKeyRotationPolicy converts a Terraform rotation_policy value to a *key_manager.KeyRotationPolicy.
func ExpandKeyRotationPolicy(v any) (*key_manager.KeyRotationPolicy, error) {
	list, ok := v.([]any)
	if !ok || len(list) == 0 {
		return nil, nil
	}

	m, ok := list[0].(map[string]any)
	if !ok {
		return nil, nil
	}

	periodStr, ok := m["rotation_period"].(string)
	if !ok || periodStr == "" {
		return nil, nil
	}

	period, err := time.ParseDuration(periodStr)
	if err != nil {
		return nil, err
	}

	return &key_manager.KeyRotationPolicy{
		RotationPeriod: scw.NewDurationFromTimeDuration(period),
	}, nil
}

// FlattenKeyRotationPolicy converts a *key_manager.KeyRotationPolicy to a []map[string]any for Terraform.
func FlattenKeyRotationPolicy(rp *key_manager.KeyRotationPolicy) []map[string]any {
	if rp == nil {
		return nil
	}

	var periodStr string

	if rp.RotationPeriod != nil {
		periodStr = rp.RotationPeriod.ToTimeDuration().String()
	}

	return []map[string]any{
		{
			"rotation_period":  periodStr,
			"next_rotation_at": types.FlattenTime(rp.NextRotationAt),
		},
	}
}
