package object

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetBucketState_ListObjects500DoesNotFailRead reproduces the issue where
// ListObjects returns HTTP 500 (InternalError) on a bucket with millions of
// delete markers at the start of its keyspace. The bucket read fails with an
// Error diagnostic even though no attribute of scaleway_object_bucket comes
// from the object listing — the result is discarded.
//
// The bucket's existence is already established earlier in setBucketState by
// GetBucketAcl and GetObjectLockConfiguration. A ListObjects failure must not
// fail the entire read.
func TestSetBucketState_ListObjects500DoesNotFailRead(t *testing.T) {
	// Clear env vars so the profile is not overridden by host config.
	for _, env := range []string{
		scw.ScwS3EndpointEnv,
		scw.ScwS3UsePathStyleEnv,
		scw.ScwAccessKeyEnv,
		scw.ScwSecretKeyEnv,
		scw.ScwDefaultProjectIDEnv,
		scw.ScwDefaultOrganizationIDEnv,
		scw.ScwDefaultRegionEnv,
		scw.ScwDefaultZoneEnv,
		"SCW_CONFIG_PATH",
	} {
		t.Setenv(env, "")
	}

	bucketName := "test-bucket-list-objects-500"

	m, err := meta.NewMetaFromProfile(t.Context(), &scw.Profile{
		AccessKey:     scw.StringPtr("SCWXXXXXXXXXXXXXXXXX"),
		SecretKey:     scw.StringPtr("866f4a9a-d058-4d3c-a39f-86930849ccc0"),
		DefaultRegion: scw.StringPtr(scw.RegionFrPar.String()),
		DefaultZone:   scw.StringPtr(scw.ZoneFrPar1.String()),
	}, nil, nil, nil, "test", &http.Client{})
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		switch {
		case q.Has("acl"):
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Owner><DisplayName>project-id:owner</DisplayName><ID>project-id:owner</ID></Owner>
  <AccessControlList>
    <Grant>
      <Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
        <ID>project-id:owner</ID><DisplayName>project-id:owner</DisplayName>
      </Grantee>
      <Permission>FULL_CONTROL</Permission>
    </Grant>
  </AccessControlList>
</AccessControlPolicy>`)

		case q.Has("object-lock"):
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<Error><Code>ObjectLockConfigurationNotFoundError</Code><Message>Object Lock configuration does not exist for this bucket</Message></Error>`)

		case q.Has("tagging"):
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<Error><Code>NoSuchTagSet</Code><Message>The TagSet does not exist</Message></Error>`)

		case q.Has("cors"):
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<Error><Code>NoSuchCORSConfiguration</Code><Message>The CORS configuration does not exist</Message></Error>`)

		case q.Has("versioning"):
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><VersioningConfiguration><Status></Status></VersioningConfiguration>`)

		case q.Has("lifecycle"):
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<Error><Code>NoSuchLifecycleConfiguration</Code><Message>The lifecycle configuration does not exist</Message></Error>`)

		default:
			// ListObjects — returns 500 InternalError, the bug trigger.
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `<Error><Code>InternalError</Code><Message>We encountered an internal error. Please try again.</Message><RequestId>test-req-id</RequestId><HostId>test-host-id</HostId></Error>`)
		}
	}))
	defer srv.Close()

	ctx := context.Background()

	s3Client, err := newS3Client(ctx, "fr-par", "SCWXXXXXXXXXXXXXXXXX", "secret-key", srv.URL, true, &http.Client{})
	require.NoError(t, err)

	d := schema.TestResourceDataRaw(t, bucketSchema(), map[string]any{
		"name": bucketName,
	})
	d.SetId(fmt.Sprintf("fr-par/%s", bucketName))

	diags := setBucketState(ctx, d, m, bucketName, scw.RegionFrPar, s3Client)

	// The ListObjects 500 must not produce an Error diagnostic.
	// No attribute of scaleway_object_bucket is derived from the object listing;
	// the result is discarded. The bucket's existence is already confirmed by
	// GetBucketAcl and GetObjectLockConfiguration earlier in the function.
	for _, d := range diags {
		assert.NotEqual(t, diag.Error, d.Severity,
			"ListObjects failure should not fail the bucket read: %s", d.Summary)
	}

	// The bucket must remain in state.
	assert.Equal(t, fmt.Sprintf("fr-par/%s", bucketName), d.Id())
}
