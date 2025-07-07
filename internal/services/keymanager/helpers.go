package keymanager

import (
	"fmt"
	"strings"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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

// NewKeyManagerAPIWithRegionAndID returns a Key Manager API client, region, and key ID from meta and a composite ID.
func NewKeyManagerAPIWithRegionAndID(m any, id string) (*key_manager.API, scw.Region, string, error) {
	region, keyID, err := ExtractRegionAndKeyID(id)
	if err != nil {
		return nil, "", "", err
	}

	client := key_manager.NewAPI(meta.ExtractScwClient(m))

	return client, region, keyID, nil
}
