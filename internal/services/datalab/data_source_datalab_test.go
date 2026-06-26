package datalab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDatalabDataSource_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_ds_by_id") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_ds_by_id"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						description        = "test datalab for data source by id"
						tags               = ["ds-test"]

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
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_ds_by_id") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_ds_by_id"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						description        = "test datalab for data source by id"
						tags               = ["ds-test"]

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}

					data "scaleway_datalab" "by_id" {
						datalab_id = scaleway_datalab.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "name", "scaleway_datalab.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "spark_version", "scaleway_datalab.main", "spark_version"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "region", "scaleway_datalab.main", "region"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "project_id", "scaleway_datalab.main", "project_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "description", "scaleway_datalab.main", "description"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "status", "scaleway_datalab.main", "status"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "has_notebook", "scaleway_datalab.main", "has_notebook"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "created_at", "scaleway_datalab.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_id", "updated_at", "scaleway_datalab.main", "updated_at"),
					resource.TestCheckResourceAttr("data.scaleway_datalab.by_id", "tags.0", "ds-test"),
					resource.TestCheckResourceAttr("data.scaleway_datalab.by_id", "tags.#", "1"),
				),
			},
		},
	})
}

func TestAccDatalabDataSource_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatalabDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_ds_by_name") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_ds_by_name"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						description        = "test datalab for data source by name"

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
				Config: fmt.Sprintf(datalabTestBaseConfig, "tf_tests_datalab_ds_by_name") + `
					resource "scaleway_datalab" "main" {
						name               = "tf_tests_datalab_ds_by_name"
						spark_version      = "4.0.0"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						description        = "test datalab for data source by name"

						main = {
							node_type = "DATALAB-SHARED-4C-8G"
						}

						worker = {
							node_type  = "DATALAB-DEDICATED2-2C-8G"
							node_count = 1
						}
					}

					data "scaleway_datalab" "by_name" {
						name   = scaleway_datalab.main.name
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatalabPresent(tt, "scaleway_datalab.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_datalab.by_name", "datalab_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "name", "scaleway_datalab.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "spark_version", "scaleway_datalab.main", "spark_version"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "region", "scaleway_datalab.main", "region"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "project_id", "scaleway_datalab.main", "project_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "description", "scaleway_datalab.main", "description"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "status", "scaleway_datalab.main", "status"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "has_notebook", "scaleway_datalab.main", "has_notebook"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "created_at", "scaleway_datalab.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_datalab.by_name", "updated_at", "scaleway_datalab.main", "updated_at"),
				),
			},
		},
	})
}
