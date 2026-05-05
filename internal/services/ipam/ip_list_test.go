package ipam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListIPAMIPs_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIPAMIPs_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}

					resource "scaleway_vpc" "main" {
					  project_id = scaleway_account_project.main.id
					  region     = "fr-par"
					  name       = "test-vpc-ipam-list"
					}

					resource "scaleway_vpc_private_network" "main" {
					  project_id = scaleway_account_project.main.id
					  vpc_id     = scaleway_vpc.main.id
					  region     = "fr-par"
					  name       = "test-pn-ipam"
					}

					resource "scaleway_ipam_ip" "ip1" {
					  project_id = scaleway_account_project.main.id
					  tags       = ["test-ipam-list"]
					  source {
						private_network_id = scaleway_vpc_private_network.main.id
					  }
					}

					resource "scaleway_ipam_ip" "ip2" {
					  project_id = scaleway_account_project.main.id
					  tags       = ["test-ipam-list", "extra"]
					  source {
						private_network_id = scaleway_vpc_private_network.main.id
					  }
					}
`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_ipam_ip" "all" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_ipam_ip.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_ipam_ip" "by_tag" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
						tags        = ["extra"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_ipam_ip.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_ipam_ip" "by_pn" {
					  provider = scaleway

					  config {
						regions            = ["fr-par"]
						project_ids        = [scaleway_account_project.main.id]
						private_network_id = scaleway_vpc_private_network.main.id
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_ipam_ip.by_pn", 2),
				},
			},
		},
	})
}
