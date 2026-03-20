package datalab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDatalabsDataSource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_basic") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_basic"
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
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_basic") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_basic"
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

					data "scaleway_datalabs" "all" {
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.all", "datalabs.#"),
				),
			},
		},
	})
}

func TestAccDatalabsDataSource_FilterByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_name") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_name"
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
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_name") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_name"
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

					data "scaleway_datalabs" "by_name" {
						name   = scaleway_datalab.main.name
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_name", "datalabs.0.id"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_name", "datalabs.0.name", "tf_tests_datalabs_ds_name"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_name", "datalabs.0.spark_version", "4.0.0"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_name", "datalabs.0.region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_name", "datalabs.0.project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_name", "datalabs.0.status"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_name", "datalabs.0.created_at"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_name", "datalabs.0.updated_at"),
				),
			},
		},
	})
}

func TestAccDatalabsDataSource_FilterByTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_tags") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_tags"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						tags               = ["ds-filter-test", "env:test"]

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
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalabs_ds_tags") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalabs_ds_tags"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						tags               = ["ds-filter-test", "env:test"]

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}

					data "scaleway_datalabs" "by_tags" {
						tags   = ["ds-filter-test"]
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalabs.by_tags", "datalabs.0.id"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_tags", "datalabs.0.name", "tf_tests_datalabs_ds_tags"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_tags", "datalabs.0.tags.#", "2"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_tags", "datalabs.0.tags.0", "ds-filter-test"),
					resource.TestCheckResourceAttr("data.scaleway_datalabs.by_tags", "datalabs.0.tags.1", "env:test"),
				),
			},
		},
	})
}
