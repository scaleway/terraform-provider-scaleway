package keymanager

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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

func newKeyManagerAPI(d *schema.ResourceData, m any) (*key_manager.API, scw.Region, error) {
	api := key_manager.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func NewKeyManagerAPIWithRegionAndID(m any, id string) (*key_manager.API, scw.Region, string, error) {
	region, keyID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	client := key_manager.NewAPI(meta.ExtractScwClient(m))

	return client, region, keyID, nil
}

func ExpandKeyUsage(usage string) *key_manager.KeyUsage {
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
		return nil, errors.New("rotation_period is required when rotation_policy block is specified")
	}

	period, err := time.ParseDuration(periodStr)
	if err != nil {
		return nil, err
	}

	policy := &key_manager.KeyRotationPolicy{
		RotationPeriod: scw.NewDurationFromTimeDuration(period),
	}

	// Handle next_rotation_at if provided
	if nextRotationStr, ok := m["next_rotation_at"].(string); ok && nextRotationStr != "" {
		nextRotation, err := time.Parse(time.RFC3339, nextRotationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid next_rotation_at format: %w", err)
		}

		policy.NextRotationAt = &nextRotation
	}

	return policy, nil
}

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
