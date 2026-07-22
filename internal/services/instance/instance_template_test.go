package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccInstanceTemplateResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isTemplateDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-basic"
						tags = [ "terraform-test", "scaleway_instance_template", "basic" ]

						server_type = "PRO2-M"
						server_tags = [ "from-template" ]
						public_ip_v4_count = 1
						public_ip_v6_count = 3
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-basic"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.1", "scaleway_instance_template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.2", "basic"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.0", "from-template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ip_v4_count", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ip_v6_count", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "security_group_id"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "placement_group_id"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "windows_rdp_ssh_key_id"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "0"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "updated_at"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_instance_template" "main" {
						tags = [ "scaleway_instance_template", "basic-step2" ]

						server_type = "POP2-HM-16C-128G"
						public_ip_v4_count = 0
						public_ip_v6_count = 0
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-HM-16C-128G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.0", "scaleway_instance_template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.1", "basic-step2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ip_v4_count", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ip_v6_count", "0"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "updated_at"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInstanceTemplateResource_Volumes(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isTemplateDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-volumes"
						tags = [ "terraform-test", "scaleway_instance_template", "volumes" ]
						server_type = "GP1-L"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						volumes = [{
							image_label = "ubuntu_noble"
						}]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "GP1-L"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.volume_type", "unknown_volume_type"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.image_label", "ubuntu_noble"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.0.name"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.0.size"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.0.perf_iops"),
				),
			},
		},
	})
}

func TestAccInstanceTemplateResource_AdditionalResources(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	templateID := ""
	sshKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFNaFderD6JUbMr6LoL7SdTaQ31gLcXwKv07Zyw0t4pq6Y8CGaeEvevS54TBR2iNJHa3hlIIUmA2qvH7Oh4v1QmMG2djWi2cD1lDEl8/8PYakaEBGh6snp3TMyhoqHOZqqKwDhPW0gJbe2vXfAgWSEzI8h1fs1D7iEkC1L/11hZjkqbUX/KduWFLyIRWdSuI3SWk4CXKRXwIkeYeSYb8AiIGY21u2z8H2J7YmhRzE85Kj/Fk4tST5gLW/IfLD4TMJjC/cZiJevETjs+XVmzTMIyU2sTQKufSQTj2qZ7RfgGwTHDoOeFvylgAdMGLZ/Un+gzeEPj9xUSPvvnbA9UPIKV4AffgtT1y5gcSWuHaqRxpUTY204mh6kq0EdVN2UsiJTgX+xnJgnOrKg6G3dkM8LSi2QtbjYbRXcuDJ9YUbUFK8M5Vo7LhMsMFb1hPtY68kbDUqD01RuMD5KhGIngCRRBZJriRQclUCJS4D3jr/Frw9ruNGh+NTIvIwdv0Y2brU= opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			isSecurityGroupDestroyed(tt),
			isPlacementGroupDestroyed(tt),
			iamchecks.CheckSSHKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "sg" {
						zone = "nl-ams-3"
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-security-group"
						tags = [ "terraform-test", "scaleway_instance_template", "additional_resources" ]
						zone = "nl-ams-3"
						server_type = "PRO2-M"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						security_group_id = scaleway_instance_security_group.sg.id
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-security-group"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "security_group_id", "scaleway_instance_security_group.sg", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_instance_placement_group" "pg" {
						zone = "pl-waw-2"
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-placement-group"
						tags = [ "terraform-test", "scaleway_instance_template", "additional_resources" ]
						zone =  "pl-waw-2"
						server_type = "PLAY2-MICRO"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						placement_group_id = substr(scaleway_instance_placement_group.pg.id, 9, -1)
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isPlacementGroupPresent(tt, "scaleway_instance_placement_group.pg"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-placement-group"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "pl-waw-2"),
					acctest.CheckResourceRawIDMatches("scaleway_instance_template.main", "placement_group_id", "scaleway_instance_placement_group.pg", "id"),
					acctest.CheckResourceIDChanged("scaleway_instance_template.main", &templateID),
				),
			},
			//{
			//	ResourceName:      "scaleway_instance_template.main",
			//	ImportState:       true,
			//	ImportStateVerify: true,   // can't verify because placement_group_id will be imported in zoned format, which conflicts with its raw form in the previous step
			//},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key" {
						name = "tf-test-acc-instance-tmpl-admin-ssh-key"
						public_key = %q
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-admin-ssh-key"
						tags = [ "terraform-test", "scaleway_instance_template", "additional_resources" ]
						zone = "fr-par-1"
						server_type = "POP2-4C-16G-WIN"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						windows_rdp_ssh_key_id = scaleway_iam_ssh_key.key.id
					}`, sshKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					iamchecks.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.key"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-admin-ssh-key"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-4C-16G-WIN"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "windows_rdp_ssh_key_id", "scaleway_iam_ssh_key.key", "id"),
					acctest.CheckResourceIDChanged("scaleway_instance_template.main", &templateID),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInstanceTemplateResource_PrivateNetworks(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			vpctestfuncs.CheckPrivateNetworkDestroy(tt),
			vpctestfuncs.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc" {
						region = "nl-ams"
					}

					resource "scaleway_vpc_private_network" "pn0" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-private-networks"
						tags = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone =  "nl-ams-1"
						server_type = "PLAY2-MICRO"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						private_networks = [
							scaleway_vpc_private_network.pn0.id,
							scaleway_vpc_private_network.pn1.id,
							# substr(scaleway_vpc_private_network.pn1.id, 7, -1),
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-private-networks"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "private_networks.0", "scaleway_vpc_private_network.pn0", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "private_networks.1", "scaleway_vpc_private_network.pn1", "id"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc" {
						region = "nl-ams"
					}

					resource "scaleway_vpc_private_network" "pn0" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-private-networks"
						tags = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone =  "nl-ams-1"
						server_type = "PLAY2-MICRO"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						private_networks = [
							scaleway_vpc_private_network.pn1.id,
							scaleway_vpc_private_network.pn2.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-private-networks"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "private_networks.0", "scaleway_vpc_private_network.pn1", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "private_networks.1", "scaleway_vpc_private_network.pn2", "id"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc" {
						region = "nl-ams"
					}

					resource "scaleway_vpc_private_network" "pn0" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn3" {
						vpc_id = scaleway_vpc.vpc.id
						tags = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-private-networks"
						tags = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone =  "nl-ams-1"
						server_type = "PLAY2-MICRO"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						private_networks = []
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-private-networks"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "0"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInstanceTemplateResource_Filesystems(t *testing.T) { // TODO: wait for fix to have 2 fs attached
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			//vpctestfuncs.CheckPrivateNetworkDestroy(tt),
			//vpctestfuncs.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_file_filesystem" "fs0" {
						size_in_gb = 25
						tags = [ "scaleway_instance_template", "filesystems", "fs0" ]
					}
					resource "scaleway_file_filesystem" "fs1" {
						size_in_gb = 30
						tags = [ "scaleway_instance_template", "filesystems", "fs1" ]
					}
					resource "scaleway_file_filesystem" "fs2" {
						size_in_gb = 40
						tags = [ "scaleway_instance_template", "filesystems", "fs2" ]
					}`,
				//Check: resource.ComposeAggregateTestCheckFunc(
				//	//testAccCheckFileSystemExists(tt, "scaleway_instance_template.main"),
				//),
			},
			{
				Config: `
					resource "scaleway_file_filesystem" "fs0" {
						size_in_gb = 25
						tags = [ "scaleway_instance_template", "filesystems", "fs0" ]
					}
					resource "scaleway_file_filesystem" "fs1" {
						size_in_gb = 30
						tags = [ "scaleway_instance_template", "filesystems", "fs1" ]
					}
					resource "scaleway_file_filesystem" "fs2" {
						size_in_gb = 40
						tags = [ "scaleway_instance_template", "filesystems", "fs2" ]
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-filesystems"
						tags = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type = "POP2-8C-32G"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						filesystem_ids = [
							scaleway_file_filesystem.fs0.id,
							scaleway_file_filesystem.fs1.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-filesystems"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-8C-32G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "1"), //2"),
					//resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "filesystem_ids.0", "scaleway_file_filesystem.fs0", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "filesystem_ids.0", "scaleway_file_filesystem.fs1", "id"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_file_filesystem" "fs0" {
						size_in_gb = 25
						tags = [ "scaleway_instance_template", "filesystems", "fs0" ]
					}
					resource "scaleway_file_filesystem" "fs1" {
						size_in_gb = 30
						tags = [ "scaleway_instance_template", "filesystems", "fs1" ]
					}
					resource "scaleway_file_filesystem" "fs2" {
						size_in_gb = 40
						tags = [ "scaleway_instance_template", "filesystems", "fs2" ]
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-filesystems"
						tags = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type = "POP2-8C-32G"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						filesystem_ids = [
							substr(scaleway_file_filesystem.fs1.id, 7, -1),
							scaleway_file_filesystem.fs2.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-filesystems"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-8C-32G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "1"), //"2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "filesystem_ids.0", "scaleway_file_filesystem.fs1", "id"),
					//resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "filesystem_ids.1", "scaleway_file_filesystem.fs2", "id"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "scaleway_file_filesystem" "fs0" {
						size_in_gb = 25
						tags = [ "scaleway_instance_template", "filesystems", "fs0" ]
					}
					resource "scaleway_file_filesystem" "fs1" {
						size_in_gb = 30
						tags = [ "scaleway_instance_template", "filesystems", "fs1" ]
					}
					resource "scaleway_file_filesystem" "fs2" {
						size_in_gb = 40
						tags = [ "scaleway_instance_template", "filesystems", "fs2" ]
					}

					resource "scaleway_instance_template" "main" {
						name = "tf-test-acc-instance-tmpl-filesystems"
						tags = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type = "POP2-8C-32G"
						public_ip_v4_count = 0
						public_ip_v6_count = 0

						filesystem_ids = []
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-filesystems"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-8C-32G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "0"),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func isTemplateDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_template" {
				continue
			}

			api := instanceV2.NewAPI(tt.Meta.ScwClient())

			zone, id, err := zonal.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetTemplate(&instanceV2.GetTemplateRequest{
				Zone:       zone,
				TemplateID: id,
			})
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			return fmt.Errorf("template (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func isTemplatePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api := instanceV2.NewAPI(tt.Meta.ScwClient())

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTemplate(&instanceV2.GetTemplateRequest{
			Zone:       zone,
			TemplateID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
