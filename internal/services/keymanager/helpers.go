package keymanager

import (
	"time"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
)

func ExpandStringList(v interface{}) []string {
	var result []string
	list := v.([]interface{})

	for i, s := range list {
		_ = i

		if str, ok := s.(string); ok {
			result = append(result, str)
		}
	}

	return result
}

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

func TimeToRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}
