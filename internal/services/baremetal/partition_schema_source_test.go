package baremetal_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	baremetalSDK "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
	baremetalchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal/testfuncs"
)

const (
	offerNameEasyPartitioning            = "EM-B220E-NVME"
	incompatibleOfferName                = "EM-L110X-SATA"
	AlternativeOfferNameEasyPartitioning = "EM-B420E-NVME"
	mountpoint                           = "/data"
)

func TestAccPartitionSchemaDataSource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	sshKeyName := "TestAccPartitionSchemaDataSource_Basic"
	serverName := "TestAccPartitionSchemaDataSource_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_baremetal_os" "my_os" {
						zone    = "%s"
						name    = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name       = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id          = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id             = data.scaleway_baremetal_os.my_os.os_id
						swap              = true
						ext_4_mountpoint  = "%s"
					}

					resource "scaleway_baremetal_server" "base" {
						name         = "%s"
						zone         = "%s"
						description  = "test a description"
						offer        = data.scaleway_baremetal_offer.my_offer.offer_id
						os           = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
						tags         = ["terraform-test", "scaleway_baremetal_server", "minimal", "edited"]
						ssh_key_ids  = [scaleway_iam_ssh_key.main.id]
					}
				`,
					Zone,
					offerNameEasyPartitioning,
					Zone,
					sshKeyName,
					SSHKeyBaremetal,
					mountpoint,
					serverName,
					Zone,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckEasyPartitioning(tt, "scaleway_baremetal_server.base", "data.scaleway_baremetal_easy_partitioning.test"),
				),
			},
		},
	})
}

func TestAccPartitionSchemaDataSource_WithoutExtraPart(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	sshKeyName := "TestAccPartitionSchemaDataSource_WithoutExtraPart"
	serverName := "TestAccPartitionSchemaDataSource_WithoutExtraPart"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_baremetal_os" "my_os" {
						zone    = "%s"
						name    = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name       = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id          = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id             = data.scaleway_baremetal_os.my_os.os_id
						swap              = true
						extra_partition  = false
					}

					resource "scaleway_baremetal_server" "base" {
						name         = "%s"
						zone         = "%s"
						description  = "test a description"
						offer        = data.scaleway_baremetal_offer.my_offer.offer_id
						os           = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
						tags         = ["terraform-test", "scaleway_baremetal_server", "minimal", "edited"]
						ssh_key_ids  = [scaleway_iam_ssh_key.main.id]
					}
				`,
					Zone,
					offerNameEasyPartitioning,
					Zone,
					sshKeyName,
					SSHKeyBaremetal,
					serverName,
					Zone,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckEasyPartitioning(tt, "scaleway_baremetal_server.base", "data.scaleway_baremetal_easy_partitioning.test"),
				),
			},
		},
	})
}

func TestAccPartitionSchemaDataSource_WithoutSwapAndExtraPart(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	sshKeyName := "TestAccPartitionSchemaDataSource_WithoutSwapAndExtraPart"
	serverName := "TestAccPartitionSchemaDataSource_WithoutSwapAndExtraPart"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_baremetal_os" "my_os" {
						zone    = "%s"
						name    = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name       = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id          = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id             = data.scaleway_baremetal_os.my_os.os_id
						swap              = false
						extra_partition  = false
					}

					resource "scaleway_baremetal_server" "base" {
						name         = "%s"
						zone         = "%s"
						description  = "test a description"
						offer        = data.scaleway_baremetal_offer.my_offer.offer_id
						os           = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
						tags         = ["terraform-test", "scaleway_baremetal_server", "minimal", "edited"]
						ssh_key_ids  = [scaleway_iam_ssh_key.main.id]
					}
				`,
					Zone,
					offerNameEasyPartitioning,
					Zone,
					sshKeyName,
					SSHKeyBaremetal,
					serverName,
					Zone,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckEasyPartitioning(tt, "scaleway_baremetal_server.base", "data.scaleway_baremetal_easy_partitioning.test"),
				),
			},
		},
	})
}

func TestAccPartitionSchemaDataSource_WithAlternateOffer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccPartitionSchemaDataSource_WithAlternateOffer"
	name := "TestAccPartitionSchemaDataSource_WithAlternateOffer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Debian"
					  version = "12 (Bookworm)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id = data.scaleway_baremetal_os.my_os.os_id
						swap = false
						ext_4_mountpoint = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`,
					Zone,
					AlternativeOfferNameEasyPartitioning,
					Zone,
					SSHKeyName,
					SSHKeyBaremetal,
					mountpoint,
					name,
					Zone,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckEasyPartitioning(tt, "scaleway_baremetal_server.base", "data.scaleway_baremetal_easy_partitioning.test"),
				),
			},
		},
	})
}

func TestAccPartitionSchemaDataSource_IncompatibleOS(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccPartitionSchemaDataSource_IncompatibleOS"
	name := "TestAccPartitionSchemaDataSource_IncompatibleOS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Windows"
					  version = "2022 64BITS standard"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}

					data "scaleway_baremetal_easy_partitioning" "test" {
						offer_id = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id = data.scaleway_baremetal_os.my_os.os_id
						swap = false
						ext_4_mountpoint = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`,
					Zone,
					incompatibleOfferName,
					Zone,
					SSHKeyName,
					SSHKeyBaremetal, mountpoint, name, Zone),
				ExpectError: regexp.MustCompile("custom partitioning is not supported with this OS"),
			},
		},
	})
}

func TestAccPartitionSchemaDataSource_IncompatibleOffer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccPartitionSchemaDataSource_IncompatibleOffer"
	name := "TestAccPartitionSchemaDataSource_IncompatibleOffer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
					data "scaleway_baremetal_offer" "my_offer" {
						zone = "%s"
						name = "%s"
					}

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
						offer_id = data.scaleway_baremetal_offer.my_offer.offer_id
						os_id = data.scaleway_baremetal_os.my_os.os_id
						swap = false
						ext_4_mountpoint = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "%s"
						description = "test a description"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
						partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}`, Zone,
					incompatibleOfferName,
					Zone,
					SSHKeyName,
					SSHKeyBaremetal,
					mountpoint,
					name,
					Zone,
				),
				ExpectError: regexp.MustCompile("Custom Partitioning not supported"),
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
