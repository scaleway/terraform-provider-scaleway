package opensearch_test

import (
	"testing"

	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
)

func TestEndpointHasOpenSearchAPI(t *testing.T) {
	t.Parallel()

	public := &searchdbapi.EndpointPublicDetails{}

	tests := []struct {
		name string
		want bool
		ep   *searchdbapi.Endpoint
	}{
		{name: "nil", ep: nil, want: false},
		{
			name: "dashboard only",
			ep: &searchdbapi.Endpoint{
				Public:   public,
				Services: []*searchdbapi.EndpointService{{Name: "ui-dashboard", Port: 5601}},
			},
			want: false,
		},
		{
			name: "public with api",
			ep: &searchdbapi.Endpoint{
				Public: public,
				Services: []*searchdbapi.EndpointService{
					{Name: "ui-dashboard", Port: 5601},
					{Name: "api-opensearch", Port: 9200},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := opensearch.EndpointHasOpenSearchAPI(tt.ep); got != tt.want {
				t.Fatalf("EndpointHasOpenSearchAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
