package datalab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	datalabSDK "github.com/scaleway/scaleway-sdk-go/api/datalab/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

const datalabTestBaseConfig = `
resource "scaleway_vpc" "main" {
	name = "%s"
}

resource "scaleway_vpc_private_network" "main" {
	vpc_id = scaleway_vpc.main.id
	region = "fr-par"
}
`

func TestAccDatalabResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_basic") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_basic"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "name", "tf_tests_datalab_basic"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "spark_version", "4.0.0"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "status", "ready"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "main.node_type", "DATALAB-SHARED-4C-8G"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "worker.node_type", "DATALAB-DEDICATED2-2C-8G"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "worker.node_count", "1"),
					resource.TestCheckResourceAttrSet("scaleway_datalab.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_datalab.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_datalab.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_datalab.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccDatalabResource_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_update") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_update"
						description        = "initial description"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						tags               = ["tag1"]

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "name", "tf_tests_datalab_update"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "description", "initial description"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "tags.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_update") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_update_renamed"
						description        = "updated description"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						tags               = ["tag1", "tag2"]

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "name", "tf_tests_datalab_update_renamed"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "description", "updated description"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "status", "ready"),
				),
			},
		},
	})
}

func TestAccDatalabResource_Import(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_import") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_import"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
				),
			},
			{
				ResourceName:      "scaleway_datalab.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatalabResource_WithWorker(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_worker") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_worker"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "name", "tf_tests_datalab_worker"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "main.node_type", "DATALAB-SHARED-4C-8G"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "worker.node_type", "DATALAB-DEDICATED2-2C-8G"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "worker.node_count", "1"),
					resource.TestCheckResourceAttrSet("scaleway_datalab.main", "status"),
				),
			},
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_worker") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_worker"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 2
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "worker.node_count", "2"),
					resource.TestCheckResourceAttr("scaleway_datalab.main", "status", "ready"),
				),
			},
		},
	})
}

func isDatalabDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_datalab" {
				continue
			}

			api := datalabSDK.NewAPI(tt.Meta.ScwClient())

			region, id, err := regional.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			dl, err := api.GetDatalab(&datalabSDK.GetDatalabRequest{
				Region:    region,
				DatalabID: id,
			})
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			if dl.Status == datalabSDK.DatalabStatusDeleted {
				continue
			}

			return fmt.Errorf("datalab (%s) still exists with status %s", rs.Primary.ID, dl.Status)
		}

		return nil
	}
}

func isDatalabPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api := datalabSDK.NewAPI(tt.Meta.ScwClient())

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDatalab(&datalabSDK.GetDatalabRequest{
			Region:    region,
			DatalabID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
