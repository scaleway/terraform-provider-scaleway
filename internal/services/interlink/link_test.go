package interlink_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	interlinkSDK "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/interlink"
)

func TestAccInterlinkLink_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInterlinkLinkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link"
						tags           = ["tag1"]
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "name", "tf-test-interlink-link"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "bandwidth_mbps", "50"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttrPair(
						"scaleway_interlink_link.main", "pop_id",
						"data.scaleway_interlink_pop.pop", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_interlink_link.main", "partner_id",
						"data.scaleway_interlink_partner.partner", "id"),
					resource.TestCheckResourceAttrSet("scaleway_interlink_link.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_interlink_link.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_interlink_link.main", "updated_at"),
				),
			},
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-updated"
						tags           = ["tag1", "tag2"]
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "name", "tf-test-interlink-link-updated"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "tags.1", "tag2"),
				),
			},
			{
				ResourceName:      "scaleway_interlink_link.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInterlinkLink_WithVPC(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInterlinkLinkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_vpc" "vpc01" {
						name = "tf-test-interlink-vpc"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-vpc"
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
						vpc_id         = scaleway_vpc.vpc01.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "name", "tf-test-interlink-link-vpc"),
					resource.TestCheckResourceAttrPair("scaleway_interlink_link.main", "vpc_id", "scaleway_vpc.vpc01", "id"),
				),
			},
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_vpc" "vpc01" {
						name = "tf-test-interlink-vpc"
					}

					resource "scaleway_vpc" "vpc02" {
						name = "tf-test-interlink-swap-target"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-vpc"
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
						vpc_id         = scaleway_vpc.vpc02.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttrPair("scaleway_interlink_link.main", "vpc_id", "scaleway_vpc.vpc02", "id"),
				),
			},
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_vpc" "vpc01" {
						name = "tf-test-interlink-vpc"
					}

					resource "scaleway_vpc" "vpc02" {
						name = "tf-test-interlink-swap-target"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-vpc"
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttr("scaleway_interlink_link.main", "vpc_id", ""),
				),
			},
			{
				ResourceName:      "scaleway_interlink_link.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInterlinkLinkExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := interlink.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetLink(&interlinkSDK.GetLinkRequest{
			LinkID: id,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInterlinkLinkDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_interlink_link" {
				continue
			}

			api, region, id, err := interlink.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			link, err := api.GetLink(&interlinkSDK.GetLinkRequest{
				LinkID: id,
				Region: region,
			})
			if err == nil && link.Status != interlinkSDK.LinkStatusDeleted {
				return fmt.Errorf("interlink link (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) && err != nil {
				return err
			}
		}

		return nil
	}
}
