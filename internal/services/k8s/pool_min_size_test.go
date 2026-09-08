package k8s_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	k8sservice "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/k8s"
)

const (
	testK8SPoolClusterID       = "11111111-1111-1111-1111-111111111111"
	testK8SPoolID              = "22222222-2222-2222-2222-222222222222"
	testK8SPoolProjectID       = "33333333-3333-3333-3333-333333333333"
	testK8SPoolAccessKey       = "SCWXXXXXXXXXXXXXXXXX"
	testK8SPoolSecretKey       = "44444444-4444-4444-4444-444444444444"
	testK8SPoolRegion          = "fr-par"
	testK8SPoolZone            = "fr-par-1"
	testK8SPoolName            = "pool"
	testK8SPoolNodeType        = "pro2_xxs"
	testK8SPoolVersion         = "1.34.1"
	testK8SPoolRootVolumeType  = "sbs_5k"
	testK8SPoolServerTypeName  = "PRO2-XXS"
	testK8SPoolMetadataPath    = "/metadata"
	testK8SPoolServerTypesPath = "/instance/v1/zones/fr-par-1/products/servers"
	testK8SPoolCreatePath      = "/k8s/v1/regions/fr-par/clusters/11111111-1111-1111-1111-111111111111/pools"
	testK8SPoolClusterPath     = "/k8s/v1/regions/fr-par/clusters/11111111-1111-1111-1111-111111111111"
	testK8SPoolReadPath        = "/k8s/v1/regions/fr-par/pools/22222222-2222-2222-2222-222222222222"
	testK8SPoolNodesPathPrefix = "/k8s/v1/regions/fr-par/clusters/11111111-1111-1111-1111-111111111111/nodes"
	testK8SPoolExpectedMinSize = float64(0)
	testK8SPoolSize            = 1
	testK8SPoolMaxSize         = 10
	testK8SPoolMaxVolumeSize   = 20000000000

	testK8SPoolIdentitySchemaDiagnostic = "Resource does not have Identity schema. Please set one in order to use Identity(). This is always a problem in the provider code."
)

func TestResourceK8SPoolCreateHonorsExplicitZeroMinSize(t *testing.T) {
	t.Parallel()

	capturedCreateRequest := false
	var capturedMinSize float64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == testK8SPoolMetadataPath:
			writeK8SPoolTestResponse(t, w, testK8SPoolMetadataResponse())
		case r.Method == http.MethodGet && r.URL.Path == testK8SPoolServerTypesPath:
			writeK8SPoolTestResponse(t, w, testK8SPoolServerTypesResponse())
		case r.Method == http.MethodGet && r.URL.Path == testK8SPoolClusterPath:
			writeK8SPoolTestResponse(t, w, testK8SPoolClusterResponse())
		case r.Method == http.MethodPost && r.URL.Path == testK8SPoolCreatePath:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			minSize, ok := body["min_size"].(float64)
			if !ok {
				http.Error(w, "request body min_size is not a number", http.StatusBadRequest)
				return
			}
			capturedCreateRequest = true
			capturedMinSize = minSize
			writeK8SPoolTestResponse(t, w, testK8SPoolResponse())
		case r.Method == http.MethodGet && r.URL.Path == testK8SPoolReadPath:
			writeK8SPoolTestResponse(t, w, testK8SPoolResponse())
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, testK8SPoolNodesPathPrefix):
			writeK8SPoolTestResponse(t, w, `{"nodes":[]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	m, err := meta.NewMetaFromProfile(context.Background(), &scw.Profile{
		AccessKey:        testStringPtr(testK8SPoolAccessKey),
		SecretKey:        testStringPtr(testK8SPoolSecretKey),
		DefaultProjectID: testStringPtr(testK8SPoolProjectID),
		DefaultRegion:    testStringPtr(testK8SPoolRegion),
		DefaultZone:      testStringPtr(testK8SPoolZone),
		APIURL:           testStringPtr(server.URL),
	}, &meta.CredentialsSource{}, "terraform-tests", server.Client())
	if err != nil {
		t.Fatalf("creating meta: %s", err)
	}

	poolResource := k8sservice.ResourcePool()
	d := schema.TestResourceDataRaw(t, poolResource.SchemaFunc(), map[string]any{
		"cluster_id":          testK8SPoolClusterID,
		"name":                testK8SPoolName,
		"node_type":           testK8SPoolNodeType,
		"size":                testK8SPoolSize,
		"min_size":            int(testK8SPoolExpectedMinSize),
		"max_size":            testK8SPoolMaxSize,
		"root_volume_type":    testK8SPoolRootVolumeType,
		"wait_for_pool_ready": false,
	})

	assertOnlyK8SPoolTestHarnessDiagnostics(t, k8sservice.ResourceK8SPoolCreate(context.Background(), d, m))

	if !capturedCreateRequest {
		t.Fatal("CreatePool request was not sent")
	}
	if capturedMinSize != testK8SPoolExpectedMinSize {
		t.Fatalf("CreatePool min_size = %v, want %v", capturedMinSize, testK8SPoolExpectedMinSize)
	}
}

func testStringPtr(value string) *string {
	return &value
}

func writeK8SPoolTestResponse(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	if _, err := fmt.Fprint(w, body); err != nil {
		t.Errorf("writing response: %s", err)
	}
}

func assertOnlyK8SPoolTestHarnessDiagnostics(t *testing.T, diags diag.Diagnostics) {
	t.Helper()

	if len(diags) == 0 {
		return
	}

	if len(diags) == 1 &&
		diags[0].Severity == diag.Error &&
		diags[0].Summary == testK8SPoolIdentitySchemaDiagnostic {
		return
	}

	t.Fatalf("creating pool returned diagnostics: %v", diags)
}

func testK8SPoolMetadataResponse() string {
	return `{"platform":"external","partition":"scw","domain":"scw.eu"}`
}

func testK8SPoolServerTypesResponse() string {
	return fmt.Sprintf(`{
		"servers":{
			%q:{
				"volumes_constraint":{
					"min_size":0,
					"max_size":%d
				}
			}
		}
	}`, testK8SPoolServerTypeName, testK8SPoolMaxVolumeSize)
}

func testK8SPoolClusterResponse() string {
	return fmt.Sprintf(`{
		"region":%q,
		"id":%q,
		"project_id":%q,
		"type":"kapsule",
		"name":"cluster",
		"status":"ready",
		"version":%q,
		"cni":"cilium",
		"tags":[],
		"created_at":"2026-01-01T00:00:00Z",
		"updated_at":"2026-01-01T00:00:00Z"
	}`, testK8SPoolRegion, testK8SPoolClusterID, testK8SPoolProjectID, testK8SPoolVersion)
}

func testK8SPoolResponse() string {
	return fmt.Sprintf(`{
		"region":%q,
		"id":%q,
		"cluster_id":%q,
		"created_at":"2026-01-01T00:00:00Z",
		"updated_at":"2026-01-01T00:00:00Z",
		"name":%q,
		"status":"ready",
		"version":%q,
		"node_type":%q,
		"autoscaling":false,
		"size":1,
		"min_size":0,
		"max_size":10,
		"container_runtime":"containerd",
		"autohealing":false,
		"tags":[],
		"kubelet_args":{},
		"upgrade_policy":{"max_unavailable":1,"max_surge":0},
		"zone":%q,
		"root_volume_type":%q,
		"root_volume_size":%d,
		"public_ip_disabled":false,
		"security_group_id":"55555555-5555-5555-5555-555555555555"
	}`, testK8SPoolRegion, testK8SPoolID, testK8SPoolClusterID, testK8SPoolName, testK8SPoolVersion, testK8SPoolNodeType, testK8SPoolZone, testK8SPoolRootVolumeType, testK8SPoolMaxVolumeSize)
}
