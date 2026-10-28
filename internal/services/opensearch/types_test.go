package opensearch_test

import (
	"testing"

	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
)

func TestEndpointHasOpenSearchAPI(t *testing.T) {
	t.Parallel()

	if opensearch.EndpointHasOpenSearchAPI(nil) {
		t.Fatal("nil endpoint: got true, want false")
	}

	public := &searchdbapi.EndpointPublicDetails{}

	dashboardOnly := &searchdbapi.Endpoint{
		Public:   public,
		Services: []*searchdbapi.EndpointService{{Name: "ui-dashboard", Port: 5601}},
	}
	if opensearch.EndpointHasOpenSearchAPI(dashboardOnly) {
		t.Fatal("dashboard only: got true, want false")
	}

	withAPI := &searchdbapi.Endpoint{
		Public: public,
		Services: []*searchdbapi.EndpointService{
			{Name: "ui-dashboard", Port: 5601},
			{Name: "api-opensearch", Port: 9200},
		},
	}
	if !opensearch.EndpointHasOpenSearchAPI(withAPI) {
		t.Fatal("public with api: got false, want true")
	}
}
