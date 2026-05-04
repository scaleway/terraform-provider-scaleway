package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListOpenSearchDeployments_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListOpenSearchDeployments_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestVersion(tt)
	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_opensearch_deployment" "main" {
					  project_id  = scaleway_account_project.main.id
					  name        = "tf-test-opensearch-list-1"
					  version     = "%s"
					  node_amount = 1
					  node_type   = "%s"
					  password    = "ThisIsASecurePassword123!"
					  tags        = ["opensearch-list-test"]
					  volume {
					    type       = "sbs_5k"
					    size_in_gb = 5
					  }
					}

					resource "scaleway_opensearch_deployment" "alt" {
					  project_id  = scaleway_account_project.main.id
					  name        = "tf-test-opensearch-list-2"
					  version     = "%s"
					  node_amount = 1
					  node_type   = "%s"
					  password    = "ThisIsASecurePassword123!"
					  depends_on  = [scaleway_opensearch_deployment.main]
					  volume {
					    type       = "sbs_5k"
					    size_in_gb = 5
					  }
					}
				`, latestVersion, nodeType, latestVersion, nodeType),
			},
			{
				Query: true,
				Config: `
					list "scaleway_opensearch_deployment" "all" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_opensearch_deployment.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_opensearch_deployment" "by_name" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "tf-test-opensearch-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_opensearch_deployment.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_opensearch_deployment" "by_tag" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["opensearch-list-test"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_opensearch_deployment.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_opensearch_deployment" "by_version" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    version     = "%s"
					  }
					}
				`, latestVersion),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_opensearch_deployment.by_version", 2),
				},
			},
		},
	})
}
