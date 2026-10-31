package vpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestDataSourceVPCRead_NoDefaultVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// The SDK may ask the API gateway for its metadata.
		if r.URL.Path == "/metadata" {
			_, _ = w.Write([]byte(`{}`))

			return
		}

		if r.Method != http.MethodGet || r.URL.Path != "/vpc/v2/regions/fr-par/vpcs" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if got := r.URL.Query().Get("is_default"); got != "true" {
			t.Errorf("expected is_default=true filter, got %q", got)
		}

		_, _ = w.Write([]byte(`{"vpcs":[],"total_count":0}`))
	}))
	defer server.Close()

	accessKey := "SCWXXXXXXXXXXXXXXXXX"
	secretKey := "11111111-1111-1111-1111-111111111111"
	projectID := "22222222-2222-2222-2222-222222222222"
	region := "fr-par"

	m, err := meta.NewMetaFromProfile(context.Background(), &scw.Profile{
		APIURL:                &server.URL,
		AccessKey:             &accessKey,
		SecretKey:             &secretKey,
		DefaultProjectID:      &projectID,
		DefaultOrganizationID: &projectID,
		DefaultRegion:         &region,
	}, nil, nil, nil, "", server.Client())
	if err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, DataSourceVPC().SchemaMap(), map[string]any{
		"is_default": true,
		"region":     "fr-par",
	})

	diags := DataSourceVPCRead(context.Background(), d, m)
	if !diags.HasError() {
		t.Fatal("expected an error when no default VPC exists")
	}

	if !strings.Contains(diags[0].Summary, "no default VPC found in region fr-par") {
		t.Fatalf("unexpected error: %s", diags[0].Summary)
	}
}
