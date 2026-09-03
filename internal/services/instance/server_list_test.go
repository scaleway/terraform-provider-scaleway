package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	identitycheck "github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest/identity"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
	instancetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccListServers_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	identityFrPar := identitycheck.Identity()
	identityNlAms := identitycheck.Identity()
	identityPlWaw := identitycheck.Identity()

	serverReapeatedConfig := `
						image = "ubuntu_noble"
						state = "stopped"
`

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsServerDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}`,
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_instance_security_group" "sg_par" {
						project_id = scaleway_account_project.main.id  # We define the SGs explicitly to make sure it get destroyed by Terraform at the end.
					}

					resource "scaleway_instance_security_group" "sg_ams" {
						project_id = scaleway_account_project.main.id  # We define the SGs explicitly to make sure it get destroyed by Terraform at the end.
						zone       = "nl-ams-1"
					}

					resource "scaleway_instance_server" "srv_par" {%[1]s
					    zone              = "fr-par-1"
						type              = "PRO2-XS"
					    project_id        = scaleway_account_project.main.id
						security_group_id = scaleway_instance_security_group.sg_par.id
					    name              = "tf-instance-list"
					    tags              = ["tag-to-look-for"]
					}

					resource "scaleway_instance_server" "srv_ams" {%[1]s
					    zone              = "nl-ams-1"
						type              = "DEV1-S"
					    project_id        = scaleway_account_project.main.id
						security_group_id = scaleway_instance_security_group.sg_ams.id
					    name              = "tf-instance-list"
					}

					resource "scaleway_instance_server" "srv_waw" {%[1]s
					    zone = "pl-waw-1"
						type = "DEV1-S"
					    name = "tf-instance-list"
					    tags = ["tag-to-look-for"]
					}
				`, serverReapeatedConfig),
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_par"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_ams"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.srv_waw"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identityFrPar.GetIdentity("scaleway_instance_server.srv_par"),
					identityNlAms.GetIdentity("scaleway_instance_server.srv_ams"),
					identityPlWaw.GetIdentity("scaleway_instance_server.srv_waw"),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "organization" {
					  provider = scaleway

					  config {
					    zones       = ["*"]
					    project_ids = ["*"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.organization", identityFrPar.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.organization", identityNlAms.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.organization", identityPlWaw.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "default_project" {
					  provider = scaleway

					  config {
					    zones = ["*"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.default_project", 1),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.default_project", identityPlWaw.Checks()),
				},
			},
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
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_project", identityFrPar.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_project", identityNlAms.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_name" {
					  provider = scaleway

					  config {
					    zones       = ["*"]
						project_ids = [scaleway_account_project.main.id]
					    name        = "tf-instance-list"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_name", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_name", identityFrPar.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_name", identityNlAms.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_server_type" {
					  provider = scaleway

					  config {
					    zones       = ["*"]
					    project_ids = ["*"]
					    server_type = "DEV1-S"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_server_type", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_server_type", identityNlAms.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_server_type", identityPlWaw.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "by_tag" {
					  provider = scaleway

					  config {
					    zones       = ["*"]
					    project_ids = ["*"]
					    tags        = ["tag-to-look-for"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.by_tag", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_tag", identityFrPar.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.by_tag", identityPlWaw.Checks()),
				},
			},
			{
				Config: `
					resource "scaleway_account_project" "main" {}`,
			},
		},
	})
}

func TestAccListServers_BySecurityGroup(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_BySecurityGroup because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	identityServerIn1 := identitycheck.Identity()
	identityServerIn2 := identitycheck.Identity()
	identityServerOut := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsServerDestroyed(tt),
			isSecurityGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "sg1" {}
					resource "scaleway_instance_security_group" "sg2" {}

					resource "scaleway_instance_server" "in_sg1" {
					    type              = "DEV1-S"
						image             = "ubuntu_noble"
						state             = "stopped"
					    name              = "sg-in-1"
						security_group_id = scaleway_instance_security_group.sg1.id
					}

					resource "scaleway_instance_server" "in_sg2" {
					    type              = "DEV1-L"
						image             = "ubuntu_noble"
						state             = "stopped"
					    name              = "sg-in-2"
						security_group_id = scaleway_instance_security_group.sg2.id
					}

					resource "scaleway_instance_server" "out" {
					    type  = "DEV1-M"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "sg-out"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_sg1"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_sg2"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.out"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identityServerIn1.GetIdentity("scaleway_instance_server.in_sg1"),
					identityServerIn2.GetIdentity("scaleway_instance_server.in_sg2"),
					identityServerOut.GetIdentity("scaleway_instance_server.out"),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "sg1" {
					  provider = scaleway

					  config {
					    zones              = ["fr-par-1"]
						security_group_ids = [scaleway_instance_security_group.sg1.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.sg1", 1),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg1", identityServerIn1.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "sg_both" {
					  provider = scaleway

					  config {
					    zones = ["fr-par-1"]
						security_group_ids = [
						  scaleway_instance_security_group.sg1.id,
						  scaleway_instance_security_group.sg2.id,
						]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.sg_both", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg_both", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg_both", identityServerIn2.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "sg_all" {
					  provider = scaleway

					  config {
					    zones              = ["fr-par-1"]
					    security_group_ids = []
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.sg_all", 3),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg_all", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg_all", identityServerIn2.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.sg_all", identityServerOut.Checks()),
				},
			},
		},
	})
}

func TestAccListServers_ByPlacementGroup(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_ByPlacementGroup because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	identityServerIn1 := identitycheck.Identity()
	identityServerIn2 := identitycheck.Identity()
	identityServerOut := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsServerDestroyed(tt),
			isPlacementGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_placement_group" "pg1" {}
					resource "scaleway_instance_placement_group" "pg2" {}

					resource "scaleway_instance_server" "in_pg1" {
					    type               = "DEV1-S"
						image              = "ubuntu_noble"
						state              = "stopped"
					    name               = "pg-in-1"
						placement_group_id = scaleway_instance_placement_group.pg1.id
					}

					resource "scaleway_instance_server" "in_pg2" {
					    type               = "DEV1-L"
						image              = "ubuntu_noble"
						state              = "stopped"
					    name               = "pg-in-2"
						placement_group_id = scaleway_instance_placement_group.pg2.id
					}

					resource "scaleway_instance_server" "out" {
					    type  = "DEV1-M"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "pg-out"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_pg1"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_pg2"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.out"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identityServerIn1.GetIdentity("scaleway_instance_server.in_pg1"),
					identityServerIn2.GetIdentity("scaleway_instance_server.in_pg2"),
					identityServerOut.GetIdentity("scaleway_instance_server.out"),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pg1" {
					  provider = scaleway

					  config {
					    zones               = ["fr-par-1"]
						placement_group_ids = [scaleway_instance_placement_group.pg1.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pg1", 1),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg1", identityServerIn1.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pg_both" {
					  provider = scaleway

					  config {
					    zones = ["fr-par-1"]
						placement_group_ids = [
						  scaleway_instance_placement_group.pg1.id,
						  scaleway_instance_placement_group.pg2.id,
						]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pg_both", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg_both", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg_both", identityServerIn2.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pg_all" {
					  provider = scaleway

					  config {
					    zones               = ["fr-par-1"]
					    placement_group_ids = []
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pg_all", 3),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg_all", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg_all", identityServerIn2.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pg_all", identityServerOut.Checks()),
				},
			},
		},
	})
}

func TestAccListServers_ByPrivateNetwork(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_ByPrivateNetwork because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	identityServerIn1 := identitycheck.Identity()
	identityServerIn2 := identitycheck.Identity()
	identityServerOut := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsServerDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc" {}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
					}

					resource "scaleway_instance_server" "in_pn1" {
					    type  = "DEV1-S"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "pn-in-1"
						private_network {
							pn_id = scaleway_vpc_private_network.pn1.id
						}
					}

					resource "scaleway_instance_server" "in_pn2" {
					    type  = "DEV1-L"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "pn-in-2"
						private_network {
							pn_id = scaleway_vpc_private_network.pn2.id
						}
					}

					resource "scaleway_instance_server" "out" {
					    type  = "DEV1-M"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "pn-out"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_pn1"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.in_pn2"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.out"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identityServerIn1.GetIdentity("scaleway_instance_server.in_pn1"),
					identityServerIn2.GetIdentity("scaleway_instance_server.in_pn2"),
					identityServerOut.GetIdentity("scaleway_instance_server.out"),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pn1" {
					  provider = scaleway

					  config {
					    zones               = ["fr-par-1"]
						private_network_ids = [scaleway_vpc_private_network.pn1.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pn1", 1),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn1", identityServerIn1.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pn_both" {
					  provider = scaleway

					  config {
					    zones = ["fr-par-1"]
						private_network_ids = [
						  scaleway_vpc_private_network.pn1.id,
						  scaleway_vpc_private_network.pn2.id,
						]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pn_both", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn_both", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn_both", identityServerIn2.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "pn_all" {
					  provider = scaleway

					  config {
					    zones               = ["fr-par-1"]
					    private_network_ids = []
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.pn_all", 3),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn_all", identityServerIn1.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn_all", identityServerIn2.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.pn_all", identityServerOut.Checks()),
				},
			},
		},
	})
}

func TestAccListServers_ByMacAddress(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListServers_ByMacAddress because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	identityServerPublic := identitycheck.Identity()
	identityServerPrivate := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsServerDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "public_ip" {}

					resource "scaleway_instance_server" "public" {
					    type   = "DEV1-S"
						image  = "ubuntu_noble"
						state  = "stopped"
					    name   = "public"
						ip_ids = [ scaleway_instance_ip.public_ip.id ]
					}

					resource "scaleway_vpc" "vpc" {}
					resource "scaleway_vpc_private_network" "pn" {
						vpc_id = scaleway_vpc.vpc.id
					}

					resource "scaleway_instance_server" "private" {
					    type  = "DEV1-L"
						image = "ubuntu_noble"
						state = "stopped"
					    name  = "private"
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.public"),
					instancetestfuncs.IsServerPresent(tt, "scaleway_instance_server.private"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identityServerPublic.GetIdentity("scaleway_instance_server.public"),
					identityServerPrivate.GetIdentity("scaleway_instance_server.private"),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "private" {
					  provider = scaleway

					  config {
					    zones         = ["fr-par-1"]
						mac_addresses = [scaleway_instance_server.private.private_network.0.mac_address]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.private", 1),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.private", identityServerPrivate.Checks()),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_instance_server" "all" {
					  provider = scaleway

					  config {
					    zones         = ["fr-par-1"]
					    mac_addresses = []
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_instance_server.all", 2),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.all", identityServerPublic.Checks()),
					identitycheck.ExpectIdentityFunc("scaleway_instance_server.all", identityServerPrivate.Checks()),
				},
			},
		},
	})
}
