package documentdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	documentdbSDK "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/documentdb"
)

func TestAccDocumentDBReadReplica_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDocumentDBInstanceDestroy(tt),
			testAccCheckDocumentDBReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "instance" {
				  name              = "test-document_db-read-replica-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_documentdb_read_replica" "replica" {
					instance_id = scaleway_documentdb_instance.instance.id
					direct_access {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBReadReplicaExists(tt, "scaleway_documentdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_read_replica.replica", "instance_id", "scaleway_documentdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
		},
	})
}

func TestAccDocumentDBReadReplica_PrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDocumentDBInstanceDestroy(tt),
			testAccCheckDocumentDBReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "instance" {
				  name              = "test-document_db-read-replica-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc_private_network" "pn" {}

				resource "scaleway_documentdb_read_replica" "replica" {
					instance_id = scaleway_documentdb_instance.instance.id
					private_network {
						private_network_id = scaleway_vpc_private_network.pn.id
						service_ip         = "10.12.1.0/20"
					}
					depends_on         = [scaleway_vpc_private_network.pn]
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBReadReplicaExists(tt, "scaleway_documentdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_read_replica.replica", "instance_id", "scaleway_documentdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.endpoint_id"),
				),
			},
		},
	})
}

func TestAccDocumentDBReadReplica_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDocumentDBInstanceDestroy(tt),
			testAccCheckDocumentDBReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "instance" {
				  name              = "test-document_db-read-replica-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_documentdb_read_replica" "replica" {
					instance_id = scaleway_documentdb_instance.instance.id
					direct_access {}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBReadReplicaExists(tt, "scaleway_documentdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_read_replica.replica", "instance_id", "scaleway_documentdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
			{
				Config: `
				resource "scaleway_documentdb_instance" "instance" {
				  name              = "test-document_db-read-replica-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc_private_network" "pn" {}

				resource "scaleway_documentdb_read_replica" "replica" {
					instance_id = scaleway_documentdb_instance.instance.id
					private_network {
						private_network_id = scaleway_vpc_private_network.pn.id
						service_ip         = "10.12.1.0/20"
					}
					depends_on         = [scaleway_vpc_private_network.pn]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBReadReplicaExists(tt, "scaleway_documentdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_read_replica.replica", "instance_id", "scaleway_documentdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_documentdb_read_replica.replica", "direct_access.#", "0"),
				),
			},
		},
	})
}

func TestAccDocumentDBReadReplica_MultipleEndpoints(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDocumentDBInstanceDestroy(tt),
			testAccCheckDocumentDBReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "instance" {
				  name              = "test-document_db-read-replica-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc_private_network" "pn" {}

				resource "scaleway_documentdb_read_replica" "replica" {
					instance_id = scaleway_documentdb_instance.instance.id
					private_network {
						private_network_id = scaleway_vpc_private_network.pn.id
						service_ip         = "10.12.1.0/20"
					}
					direct_access {}
					depends_on         = [scaleway_vpc_private_network.pn]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBReadReplicaExists(tt, "scaleway_documentdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_read_replica.replica", "instance_id", "scaleway_documentdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_documentdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
		},
	})
}

func testAccCheckDocumentDBReadReplicaExists(tt *acctest.TestTools, readReplica string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		readReplicaResource, ok := state.RootModule().Resources[readReplica]
		if !ok {
			return fmt.Errorf("resource not found: %s", readReplica)
		}

		api, region, id, err := documentdb.NewAPIWithRegionAndID(tt.Meta, readReplicaResource.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetReadReplica(&documentdbSDK.GetReadReplicaRequest{
			Region:        region,
			ReadReplicaID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDocumentDBReadReplicaDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_documentdb_read_replica" {
				continue
			}

			api, region, id, err := documentdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetReadReplica(&documentdbSDK.GetReadReplicaRequest{
				ReadReplicaID: id,
				Region:        region,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("read_replica (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
