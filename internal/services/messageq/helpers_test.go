package messageq_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	messageqSDK "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/messageq"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/stretchr/testify/assert"
)

const deploymentTestUserName = "my_initial_user"

func fetchLatestVersion(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := messageq.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&messageqSDK.ListVersionsRequest{
		Region: scw.RegionFrPar,
	}, scw.WithContext(tt.T.Context()), scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch messageq versions: %s", err)
	}

	for _, version := range versionsResp.Versions {
		if !version.Disabled {
			return version.Version
		}
	}

	tt.T.Fatal("no messageq versions available")

	return ""
}

func fetchEligibleNodeType(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := messageq.NewAPI(tt.Meta)

	nodeTypesResp, err := api.ListNodeTypes(&messageqSDK.ListNodeTypesRequest{
		Region: scw.RegionFrPar,
	}, scw.WithContext(tt.T.Context()), scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch messageq node types: %s", err)
	}

	nodeType := selectNodeType(nodeTypesResp.NodeTypes)
	if nodeType == "" {
		tt.T.Fatalf("no enabled MessageQ node type in %s has stock status available or low_stock", scw.RegionFrPar)
	}

	return nodeType
}

func selectNodeType(nodeTypes []*messageqSDK.NodeType) string {
	for _, stockStatus := range []messageqSDK.NodeTypeStockStatus{
		messageqSDK.NodeTypeStockStatusAvailable,
		messageqSDK.NodeTypeStockStatusLowStock,
	} {
		for _, nodeType := range nodeTypes {
			if !nodeType.Disabled && nodeType.StockStatus == stockStatus {
				return nodeType.Name
			}
		}
	}

	return ""
}

func TestSelectNodeType(t *testing.T) {
	t.Parallel()

	available := &messageqSDK.NodeType{Name: "available", StockStatus: messageqSDK.NodeTypeStockStatusAvailable}
	lowStock := &messageqSDK.NodeType{Name: "low-stock", StockStatus: messageqSDK.NodeTypeStockStatusLowStock}
	disabled := &messageqSDK.NodeType{Name: "disabled", StockStatus: messageqSDK.NodeTypeStockStatusAvailable, Disabled: true}
	outOfStock := &messageqSDK.NodeType{Name: "out-of-stock", StockStatus: messageqSDK.NodeTypeStockStatusOutOfStock}

	tests := []struct {
		name      string
		want      string
		nodeTypes []*messageqSDK.NodeType
	}{
		{
			name:      "prefer available",
			want:      available.Name,
			nodeTypes: []*messageqSDK.NodeType{lowStock, available},
		},
		{
			name:      "fall back to low stock",
			want:      lowStock.Name,
			nodeTypes: []*messageqSDK.NodeType{outOfStock, disabled, lowStock},
		},
		{
			name:      "no eligible node type",
			want:      "",
			nodeTypes: []*messageqSDK.NodeType{disabled, outOfStock},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, selectNodeType(tt.nodeTypes))
		})
	}
}

func isDeploymentPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := messageq.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		err = transport.RetryOnStatus(tt.T.Context(), transport.IAMPropagation, func(ctx context.Context) error {
			_, err = api.GetDeployment(&messageqSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(ctx))

			return err
		})

		return err
	}
}
