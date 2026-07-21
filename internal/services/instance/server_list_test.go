package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
	instancetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccListServers_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverReapeatedConfig := `
						type = "DEV1-S"
						image = "ubuntu_noble"
						state = "stopped"
`

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			accounttestfuncs.IsProjectDestroyed(tt),
			instancetestfuncs.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}`,
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_instance_server" "srv_par" {%[1]s
					    zone       = "fr-par-1"
					    project_id = scaleway_account_project.main.id
					    name       = "tf-instance-list-par-1"
					    tags       = ["tag-to-look-for"]
					}

					resource "scaleway_instance_server" "srv_ams" {%[1]s
					    zone       = "nl-ams-1"
					    project_id = scaleway_account_project.main.id
					    name       = "tf-instance-list-ams-1"
					}

					resource "scaleway_instance_server" "srv_waw" {%[1]s
					    zone = "pl-waw-1"
					    name = "tf-conn-list"
					    tags = ["tag-to-look-for"]
					}
				`, serverReapeatedConfig),
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_par"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_ams"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_waw"),
				),
			},
			//{
			//	Query: true,
			//	Config: `
			//		list "scaleway_instance_server" "all" {
			//		  provider = scaleway
			//
			//		  config {
			//		    zones = ["*"]
			//		  }
			//		}
			//	`,
			//	QueryResultChecks: []querycheck.QueryResultCheck{
			//		querycheck.ExpectLength("list.scaleway_instance_server.all", 3),
			//	},  // TODO: for now, only returns 1 element, the pl-waw server, because my default project id gets added to the request, therefore excluding the fr-par and nl-ams servers because they have another project id.
			//},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_project" {
					  provider = scaleway

					  config {
					    zones       = ["*"]
					    project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_project", 2),
					//querycheck.ExpectIdentity("list.scaleway_instance_server.by_project", map[string]knownvalue.Check{  // does not work for List Resources, it checks the identity of the List Resource and not the underlying resources contained in the list
					//	"zone": knownvalue.StringExact("fr-par-1"),
					//}),
					//querycheck.ExpectIdentity("list.scaleway_instance_server.by_project", map[string]knownvalue.Check{
					//	"zone": knownvalue.StringExact("nl-ams-1"),
					//}),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_name" {
					  provider = scaleway

					  config {
					    zones = ["fr-par"]
					    name  = "tf-instance-list"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_name", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_tag" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["tag-to-look-for"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_tag", 2),
				},
			},
			{
				Config: `
					resource "scaleway_account_project" "main" {}`,
			},
		},
	})
}
