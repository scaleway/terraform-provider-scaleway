package keymanager

import (
	"time"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
)

func ExpandStringList(v interface{}) []string {
	if v == nil {
		return nil
	}
	list := v.([]interface{})
	result := make([]string, len(list))
	for i, s := range list {
		result[i] = s.(string)
	}
	return result
}

func FlattenUsage(u *key_manager.KeyUsage) string {
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

func FlattenTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
