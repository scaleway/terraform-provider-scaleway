package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_mnq_namespace", &resource.Sweeper{
		Name: "scaleway_mnq_namespace",
		F:    testSweepMNQNamespace,
	})
}

func testSweepMNQNamespace(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		mnqAPI := mnq.NewAPI(scwClient)
		l.Infof("sweeper: destroying the mnq namespaces in (%s)", region)
		listNamespaces, err := mnqAPI.ListNamespaces(
			&mnq.ListNamespacesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing mnq namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			err = mnqAPI.DeleteNamespace(&mnq.DeleteNamespaceRequest{
				NamespaceID: ns.ID,
				Region:      region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)
				return fmt.Errorf("error deleting mqn namespace in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayMNQNamespace_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQnNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  name     = "test-mnq-ns-basic-1"
					  protocol = "nats"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNamespaceExists(tt, "scaleway_mnq_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_namespace.main", "protocol", "nats"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "region"),
				),
			},
			{
				Config: `
					resource scaleway_mnq_namespace main {
					  name     = "test-mnq-ns-basic-2"
					  protocol = "nats"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNamespaceExists(tt, "scaleway_mnq_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_namespace.main", "name", "test-mnq-ns-basic-2"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  name     = "test-mnq-ns-basic-1"
					  protocol = "nats"
					  region   = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNamespaceExists(tt, "scaleway_mnq_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_namespace.main", "name", "test-mnq-ns-basic-1"),
					resource.TestCheckResourceAttr("scaleway_mnq_namespace.main", "region", "fr-par"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  protocol = "sqs_sns"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNamespaceExists(tt, "scaleway_mnq_namespace.main"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "name"),
					resource.TestCheckResourceAttr("scaleway_mnq_namespace.main", "protocol", "sqs_sns"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQNamespaceExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetNamespace(&mnq.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQnNamespaceDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_namespace" {
				continue
			}

			api, region, id, err := mnqAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteNamespace(&mnq.DeleteNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("mnq namespace (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
