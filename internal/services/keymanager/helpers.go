package keymanager

import (
	"fmt"
	"strings"
	"time"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func ExpandStringList(v any) []string {
	var result []string

	list := v.([]any)

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

// ExtractRegionAndKeyID parses an ID of the form "region/key_id" and returns the region and key ID.
func ExtractRegionAndKeyID(id string) (scw.Region, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected ID format (%s), expected region/key_id", id)
	}

	return scw.Region(parts[0]), parts[1], nil
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
