package applesilicon_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	applesiliconSDK "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/applesilicon"
)

var (
	devOSID     = "cafecafe-5018-4dcd-bd08-35f031b0ac3e"
	githubUrl   = os.Getenv("GITHUB_URL_AS")
	githubToken = os.Getenv("GITHUB_TOKEN_AS")
)

func TestAccServer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "TestAccServerBasic"
						type = "M4-M"
						public_bandwidth = 1000000000
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerBasic"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M4-M"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "public_bandwidth", "1000000000"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
				),
			},
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "TestAccServerBasic"
						type = "M4-M"
						public_bandwidth = 2000000000
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerBasic"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M4-M"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "public_bandwidth", "2000000000"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
				),
			},
			{
				ResourceName:      "scaleway_apple_silicon_server.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServer_Runner(t *testing.T) {
	t.Skip("can not register this cassette for security issue")
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_apple_silicon_runner" "main" {
						name       = "TestAccRunnerGithub"
						ci_provider   = "github"
						url        = "%s"
						token      = "%s"
					}

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerRunner"
						type = "M2-L"
						public_bandwidth = 1000000000
						os_id = "%s"
						runner_ids = [scaleway_apple_silicon_runner.main.id]
					}
				`, githubUrl, githubToken, devOSID),
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerRunner"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-L"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "public_bandwidth", "1000000000"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "os_id", devOSID),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "runner_ids.0"),
				),
			},
		},
	})
}

func TestAccServer_EnableDisabledVPC(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableDisableVPC"
						type = "M2-M"
						enable_vpc = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableDisableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "vpc_status", "vpc_enabled"),
				),
			},
			{
				Config: `

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableDisableVPC"
						type = "M2-M"
						enable_vpc = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableDisableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "vpc_status", "vpc_disabled"),
				),
			},
		},
	})
}

func TestAccServer_EnableVPC(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "TestAccServerEnableVPC"
					}
					
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "TestAccServerEnableVPC"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableVPC"
						type = "M2-M"
						enable_vpc = true
						private_network {
						  id = scaleway_vpc_private_network.pn01.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "commitment", "duration_24h"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
					resource.TestCheckResourceAttrPair("scaleway_apple_silicon_server.main", "private_network.0.id", "scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "vpc_status", "vpc_enabled"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.0.address"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.1.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.1.address"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "TestAccServerEnableVPC"
					}
					
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "TestAccServerEnableVPC"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "TestAccServerEnableVPCTwo"
					}
					
					resource "scaleway_vpc_private_network" "pn02" {
					  name = "TestAccServerEnableVPCNumbertwo"
					  vpc_id = scaleway_vpc.vpc02.id
					}

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableVPC"
						type = "M2-M"
						enable_vpc = true
						private_network {
						  id = scaleway_vpc_private_network.pn01.id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn02.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
					resource.TestCheckResourceAttrPair("scaleway_apple_silicon_server.main", "private_network.0.id", "scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_apple_silicon_server.main", "private_network.1.id", "scaleway_vpc_private_network.pn02", "id"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "vpc_status", "vpc_enabled"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.0.address"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.1.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.1.address"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.2.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.2.address"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.3.id"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "private_ips.3.address"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "TestAccServerEnableVPC"
					}
					
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "TestAccServerEnableVPC"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "TestAccServerEnableVPCTwo"
					}
					
					resource "scaleway_vpc_private_network" "pn02" {
					  name = "TestAccServerEnableVPCNumbertwo"
					  vpc_id = scaleway_vpc.vpc02.id
					}

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableVPC"
						type = "M2-M"
						enable_vpc = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "vpc_status", "vpc_disabled"),
				),
			},
		},
	})
}

func TestAccServer_Commitment(t *testing.T) {
	t.Skip("can not delete server at the time")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableDisableVPC"
						type = "M2-M"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableDisableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "commitment", "duration_24h"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
				),
			},
			{
				Config: `

					resource scaleway_apple_silicon_server main {
						name = "TestAccServerEnableDisableVPC"
						type = "M2-M"
						commitment = "renewed_monthly"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "TestAccServerEnableDisableVPC"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", "M2-M"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "commitment", "renewed_monthly"),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
				),
			},
		},
	})
}

func isServerPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		asAPI, zone, ID, err := applesilicon.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = asAPI.GetServer(&applesiliconSDK.GetServerRequest{
			ServerID: ID,
			Zone:     zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isServerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_apple_silicon_server" {
				continue
			}

			asAPI, zone, ID, err := applesilicon.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = asAPI.GetServer(&applesiliconSDK.GetServerRequest{
				ServerID: ID,
				Zone:     zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("server (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}
