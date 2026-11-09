package acctest_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestCheckSRN exercises CheckSRN against the SRN specification.
//
// SRN pattern: srn://product-slug.domain/(locality_type/{locality}/)resource-type/resource-id/...
// See internal spec: namespaces (product-slug.domain), per-locality segments,
// non-UUID ids (object storage, version numbers), and sub-resources.
func TestCheckSRN(t *testing.T) {
	const uuid = "25f6c309-a2c6-4fb2-8d16-57960c3fb9df"

	tests := []struct {
		name         string
		service      string
		resourceType string
		value        string
		wantErr      bool
	}{
		// --- Valid SRNs (spec examples) ---
		{
			name:         "simple resource without locality",
			service:      "iam",
			resourceType: "policies",
			value:        "srn://iam.scw.eu/policies/" + uuid,
		},
		{
			name:         "zone locality segment",
			service:      "instance",
			resourceType: "servers",
			value:        "srn://instance.scw.eu/zones/fr-par-1/servers/" + uuid,
		},
		{
			name:         "region locality segment",
			service:      "instance",
			resourceType: "servers",
			value:        "srn://instance.scw.eu/regions/fr-par/servers/" + uuid,
		},
		{
			name:         "k8s cluster with region locality",
			service:      "k8s",
			resourceType: "clusters",
			value:        "srn://k8s.scw.eu/regions/fr-par/clusters/" + uuid,
		},
		{
			name:         "multiple locality patterns supported for same resource type",
			service:      "instance",
			resourceType: "servers",
			value:        "srn://instance.scw.eu/zones/fr-par-1/servers/" + uuid,
		},
		{
			name:         "sub-resource with non-UUID id (version number)",
			service:      "secret-manager",
			resourceType: "versions",
			value:        "srn://secret-manager.scw.eu/regions/fr-par/secrets/" + uuid + "/versions/5",
		},
		{
			name:         "object storage bucket (non-UUID id)",
			service:      "object",
			resourceType: "buckets",
			value:        "srn://object.scw.eu/buckets/bucket-name",
		},
		{
			name:         "object storage object path (resource type check skipped)",
			service:      "object",
			resourceType: "",
			value:        "srn://object.scw.eu/buckets/bucket-name/objects/path/to/my/object",
		},
		{
			name:         "internal partition domain with multiple dots",
			service:      "instance",
			resourceType: "servers",
			value:        "srn://instance.0x55.fr/zones/fr-par-1/servers/" + uuid,
		},
		{
			name:         "staging partition domain",
			service:      "object",
			resourceType: "buckets",
			value:        "srn://object.scw-stg.eu/buckets/bucket-name",
		},
		{
			name:         "empty service skips service check",
			service:      "",
			resourceType: "clusters",
			value:        "srn://k8s.scw.eu/regions/fr-par/clusters/" + uuid,
		},
		{
			name:         "empty resource type skips resource type check",
			service:      "iam",
			resourceType: "",
			value:        "srn://iam.scw.eu/whatever/" + uuid,
		},

		// --- Invalid SRNs ---
		{
			name:         "scheme mismatch: https",
			service:      "iam",
			resourceType: "policies",
			value:        "https://iam.scw.eu/policies/" + uuid,
			wantErr:      true,
		},
		{
			// An SRN can be *used as* a SPIFFE ID, but a spiffe:// URI is not
			// an SRN; CheckSRN is strict about the srn scheme.
			name:         "scheme mismatch: spiffe is not srn",
			service:      "instance",
			resourceType: "servers",
			value:        "spiffe://instance.scw.eu/zones/fr-par-1/servers/" + uuid,
			wantErr:      true,
		},
		{
			name:         "host missing domain (no dot)",
			service:      "iam",
			resourceType: "policies",
			value:        "srn://iam/policies/" + uuid,
			wantErr:      true,
		},
		{
			name:         "service mismatch",
			service:      "iam",
			resourceType: "policies",
			value:        "srn://k8s.scw.eu/policies/" + uuid,
			wantErr:      true,
		},
		{
			name:         "resource type mismatch",
			service:      "iam",
			resourceType: "policies",
			value:        "srn://iam.scw.eu/users/" + uuid,
			wantErr:      true,
		},
		{
			name:         "empty path",
			service:      "iam",
			resourceType: "policies",
			value:        "srn://iam.scw.eu/",
			wantErr:      true,
		},
		{
			name:         "locality without resource type and id",
			service:      "instance",
			resourceType: "servers",
			value:        "srn://instance.scw.eu/zones/fr-par-1/",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := acctest.CheckSRN(tt.service, tt.resourceType)
			err := check(tt.value)
			switch {
			case tt.wantErr && err == nil:
				t.Fatalf("CheckSRN(%q, %q)(%q) expected error, got nil", tt.service, tt.resourceType, tt.value)
			case !tt.wantErr && err != nil:
				t.Fatalf("CheckSRN(%q, %q)(%q) expected no error, got %v", tt.service, tt.resourceType, tt.value, err)
			}
		})
	}
}
