package baremetal_test

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	baremetalSDK "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const (
	offerIDPartitioning = "a60ae97c-268c-40cb-af5f-dd276e917ed7"
	osID                = "7d1914e1-f4ab-47fc-bd8c-b3a23143e87a"
	incompatibleOsID    = "4aff4d9d-b1f4-44b0-ab6f-e4711ac11711"
	incompatibleOfferId = "a204136d-656b-44b7-9735-88ca2f62cb1f"
)

func TestAccDataSourceEasyParitioning_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	//mountpoint := "/hello"
	SSHKeyName := "TestAccServer_Basic"
	name := "TestAccServer_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id = "%s"
						os_id = "%s"
						swap = false
						ext_4_mountpoint = "/hello"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = "%s"
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`, Zone, SSHKeyName, SSHKeyBaremetal, offerIDPartitioning, osID, name, Zone, offerIDPartitioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckEasyPartitioning(tt, "scaleway_baremetal_server.base", "data.scaleway_baremetal_easy_partitioning.test"),
				),
			},
		},
	})
}

func TestAccDataSourceEasyParitioning_NotCompatibleOS(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	//mountpoint := "/hello"
	SSHKeyName := "TestAccServer_Basic"
	name := "TestAccServer_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id = "%s"
						os_id = "%s"
						swap = false
						ext_4_mountpoint = "/hello"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = "%s"
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`, Zone, SSHKeyName, SSHKeyBaremetal, offerIDPartitioning, incompatibleOsID, name, Zone, offerIDPartitioning),
				ExpectError: regexp.MustCompile("custom partitioning is not supported with this OS"),
			},
		},
	})
}

func TestAccDataSourceEasyParitioning_NotCompatibleOffer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	//mountpoint := "/hello"
	SSHKeyName := "TestAccServer_Basic"
	name := "TestAccServer_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id = "%s"
						os_id = "%s"
						swap = false
						ext_4_mountpoint = "/hello"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = "%s"
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`, Zone, SSHKeyName, SSHKeyBaremetal, offerIDPartitioning, osID, name, Zone, offerIDPartitioning),
				ExpectError: regexp.MustCompile(".*offer_id not compatible.*"),
			},
		},
	})
}

func testAccCheckEasyPartitioning(tt *acctest.TestTools, serverName, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serverName]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverName)
		}

		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		partitionJSON, ok := ds.Primary.Attributes["json_partition"]
		if !ok {
			return fmt.Errorf("attribute json_partition not found in %s", dataSourceName)
		}

		baremetalAPI, zonedID, err := baremetal.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to create API: %w", err)
		}

		server, err := baremetalAPI.GetServer(&baremetalSDK.GetServerRequest{
			ServerID: zonedID.ID,
			Zone:     zonedID.Zone,
		})
		if err != nil {
			return fmt.Errorf("failed to get server: %w", err)
		}

		if server.Install == nil || server.Install.PartitioningSchema == nil {
			return fmt.Errorf("server %s has no partitioning schema", serverName)
		}

		var expectedSchema baremetalSDK.Schema
		err = json.Unmarshal([]byte(partitionJSON), &expectedSchema)
		if err != nil {
			return fmt.Errorf("failed to unmarshal partitionJSON: %w", err)
		}

		if !reflect.DeepEqual(&expectedSchema, server.Install.PartitioningSchema) {
			expectedStr, _ := json.MarshalIndent(expectedSchema, "", "  ")
			actualStr, _ := json.MarshalIndent(server.Install.PartitioningSchema, "", "  ")
			return fmt.Errorf("partitioning schema mismatch:\nExpected:\n%s\nActual:\n%s", expectedStr, actualStr)
		}

		return nil
	}
}
