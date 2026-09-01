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
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	filetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/file/testfuncs"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccInstanceTemplateResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	// For now, we need to explicitly define the project_id for the CI's Acceptance Tests to pass.
	// We should find a more appropriate solution in the future.
	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		projectID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isTemplateDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id = %q
						tags = [ "terraform-test", "scaleway_instance_template", "basic" ]

						server_type = "PRO2-M"
						server_tags = [ "from-template" ]
						public_ipv4_count = 1
						public_ipv6_count = 3
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.1", "scaleway_instance_template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.2", "basic"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.0", "from-template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv4_count", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv6_count", "3"),
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
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id = %q
						name = "tf-test-acc-instance-tmpl-basic"
						tags = [ "scaleway_instance_template", "basic-step2" ]

						server_type = "POP2-HM-16C-128G"
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-basic"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-HM-16C-128G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.0", "scaleway_instance_template"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "tags.1", "basic-step2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv4_count", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv6_count", "0"),
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

	templateID := ""

	// For now, we need to explicitly define the project_id for the CI's Acceptance Tests to pass.
	// We should find a more appropriate solution in the future.
	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		projectID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isTemplateDestroyed(tt),
			blocktestfuncs.IsSnapshotDestroyed(tt),
			blocktestfuncs.IsVolumeDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id  = %q
						name        = "tf-test-acc-instance-tmpl-volumes"
						tags        = [ "terraform-test", "scaleway_instance_template", "volumes" ]
						server_type = "GP1-L"
						zone        = "fr-par-2"

						volumes = [{
							volume_type = "l_ssd"
							image_label = "ubuntu_noble"
							size_in_gb  = 20
							tags        = [ "local", "volume" ]
						}]
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "name", "tf-test-acc-instance-tmpl-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "GP1-L"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.image_label", "ubuntu_noble"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.size_in_gb", "20"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.0", "local"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.1", "volume"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.0.name"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "volumes.0.perf_iops"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_block_volume" "sbs" {
						size_in_gb = 15
						iops       = 5000
						zone       = "fr-par-2"
					}
					resource "scaleway_block_snapshot" "snap" {
						volume_id = scaleway_block_volume.sbs.id
						zone      = "fr-par-2"
					}

					resource "scaleway_instance_template" "main" {
						project_id  = %q
						name        = "tf-test-acc-instance-tmpl-volumes"
						tags        = [ "terraform-test", "scaleway_instance_template", "volumes" ]
						server_type = "GP1-L"
						zone        = "fr-par-2"

						volumes = [{
							volume_type = "l_ssd"
							image_label = "debian_trixie"
							size_in_gb  = 25
							tags        = [ "local" ]
							name        = "root-volume"
						},{
							volume_type      = "sbs"
							base_snapshot_id = scaleway_block_snapshot.snap.id
							size_in_gb       = 30
							perf_iops        = 15000
						}]
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.snap"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.image_label", "debian_trixie"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.size_in_gb", "25"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.0", "local"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.name", "root-volume"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "volumes.0.perf_iops"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.volume_type", "sbs"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "volumes.1.base_snapshot_id", "scaleway_block_snapshot.snap", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.size_in_gb", "30"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.perf_iops", "15000"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.tags.#", "0"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.1.name"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_block_volume" "sbs" {
						size_in_gb = 15
						iops       = 5000
						zone       = "fr-par-2"
					}
					resource "scaleway_block_snapshot" "snap" {
						volume_id = scaleway_block_volume.sbs.id
						zone      = "fr-par-2"
					}

					resource "scaleway_instance_template" "main" {
						project_id  = %q
						name        = "tf-test-acc-instance-tmpl-volumes"
						tags        = [ "terraform-test", "scaleway_instance_template", "volumes" ]
						server_type = "L40S-1-48G"
						zone        = "fr-par-2"

						volumes = [{
							volume_type = "sbs"
							image_label = "ubuntu_noble_gpu_os_13_nvidia"
							size_in_gb  = 30
							perf_iops   = 15000
						},{
							volume_type = "scratch"
							size_in_gb  = 300
						}]
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "L40S-1-48G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.volume_type", "sbs"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.image_label", "ubuntu_noble_gpu_os_13_nvidia"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.size_in_gb", "30"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.perf_iops", "15000"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.0.tags.#", "0"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.0.name"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.volume_type", "scratch"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "volumes.1.image_label"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "volumes.1.base_snapshot_id"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "volumes.1.size_in_gb", "300"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "volumes.1.name"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "volumes.1.perf_iops"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
		},
	})
}

func TestAccInstanceTemplateResource_SecurityGroupID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	templateID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			isSecurityGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "sg1" {
						zone = "nl-ams-3"
						name = "first-sg"
						tags = [ "terraform-test", "scaleway_instance_template", "security_group1" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_security_group.sg1.project_id
						name        = "tf-test-acc-instance-tmpl-security-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "security_group" ]
						zone        = "nl-ams-3"
						server_type = "PRO2-M"

						security_group_id = scaleway_instance_security_group.sg1.id
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "security_group_id", "scaleway_instance_security_group.sg1", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "project_id"),
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
					resource "scaleway_instance_security_group" "sg1" {
						zone = "nl-ams-3"
						name = "first-sg"
						tags = [ "terraform-test", "scaleway_instance_template", "security_group1" ]
					}

					resource "scaleway_instance_security_group" "sg2" {
						zone = "nl-ams-3"
						name = "second-sg"
						tags = [ "terraform-test", "scaleway_instance_template", "security_group2" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_security_group.sg1.project_id
						name        = "tf-test-acc-instance-tmpl-security-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "security_group" ]
						zone        = "nl-ams-3"
						server_type = "PRO2-M"

						security_group_id = scaleway_instance_security_group.sg2.id
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "security_group_id", "scaleway_instance_security_group.sg2", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "project_id"),
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
					resource "scaleway_instance_security_group" "sg1" {
						zone = "nl-ams-3"
						name = "first-sg"
						tags = [ "terraform-test", "scaleway_instance_template", "security_group1" ]
					}

					resource "scaleway_instance_security_group" "sg2" {
						zone = "nl-ams-3"
						name = "second-sg"
						tags = [ "terraform-test", "scaleway_instance_template", "security_group2" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_security_group.sg1.project_id
						name        = "tf-test-acc-instance-tmpl-security-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "security_group" ]
						zone        = "nl-ams-3"
						server_type = "PRO2-M"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-3"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PRO2-M"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "security_group_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_template.main", "project_id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
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

func TestAccInstanceTemplateResource_PlacementGroupID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	templateID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			isPlacementGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_placement_group" "pg1" {
						zone = "pl-waw-2"
						name = "first-pg"
						tags = [ "terraform-test", "scaleway_instance_template", "placement_group1" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_placement_group.pg1.project_id
						name        = "tf-test-acc-instance-tmpl-placement-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "placement_group" ]
						zone        = "pl-waw-2"
						server_type = "PLAY2-MICRO"

						placement_group_id = substr(scaleway_instance_placement_group.pg1.id, 9, -1)
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isPlacementGroupPresent(tt, "scaleway_instance_placement_group.pg1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "pl-waw-2"),
					acctest.CheckResourceRawIDMatches("scaleway_instance_template.main", "placement_group_id", "scaleway_instance_placement_group.pg1", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_placement_group" "pg1" {
						zone = "pl-waw-2"
						name = "first-pg"
						tags = [ "terraform-test", "scaleway_instance_template", "placement_group1" ]
					}

					resource "scaleway_instance_placement_group" "pg2" {
						zone = "pl-waw-2"
						name = "second-pg"
						tags = [ "terraform-test", "scaleway_instance_template", "placement_group2" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_placement_group.pg1.project_id
						name        = "tf-test-acc-instance-tmpl-placement-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "placement_group" ]
						zone        = "pl-waw-2"
						server_type = "PLAY2-MICRO"

						placement_group_id = scaleway_instance_placement_group.pg2.id
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isPlacementGroupPresent(tt, "scaleway_instance_placement_group.pg2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "pl-waw-2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "placement_group_id", "scaleway_instance_placement_group.pg2", "id"),
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
					resource "scaleway_instance_placement_group" "pg1" {
						zone = "pl-waw-2"
						name = "first-pg"
						tags = [ "terraform-test", "scaleway_instance_template", "placement_group1" ]
					}

					resource "scaleway_instance_placement_group" "pg2" {
						zone = "pl-waw-2"
						name = "second-pg"
						tags = [ "terraform-test", "scaleway_instance_template", "placement_group2" ]
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_instance_placement_group.pg1.project_id
						name        = "tf-test-acc-instance-tmpl-placement-group"
						tags        = [ "terraform-test", "scaleway_instance_template", "placement_group" ]
						zone        = "pl-waw-2"
						server_type = "PLAY2-MICRO"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					isPlacementGroupPresent(tt, "scaleway_instance_placement_group.pg2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "pl-waw-2"),
					resource.TestCheckNoResourceAttr("scaleway_instance_template.main", "placement_group_id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
		},
	})
}

func TestAccInstanceTemplateResource_WindowsRdpSshKeyID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	templateID := ""
	sshKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFNaFderD6JUbMr6LoL7SdTaQ31gLcXwKv07Zyw0t4pq6Y8CGaeEvevS54TBR2iNJHa3hlIIUmA2qvH7Oh4v1QmMG2djWi2cD1lDEl8/8PYakaEBGh6snp3TMyhoqHOZqqKwDhPW0gJbe2vXfAgWSEzI8h1fs1D7iEkC1L/11hZjkqbUX/KduWFLyIRWdSuI3SWk4CXKRXwIkeYeSYb8AiIGY21u2z8H2J7YmhRzE85Kj/Fk4tST5gLW/IfLD4TMJjC/cZiJevETjs+XVmzTMIyU2sTQKufSQTj2qZ7RfgGwTHDoOeFvylgAdMGLZ/Un+gzeEPj9xUSPvvnbA9UPIKV4AffgtT1y5gcSWuHaqRxpUTY204mh6kq0EdVN2UsiJTgX+xnJgnOrKg6G3dkM8LSi2QtbjYbRXcuDJ9YUbUFK8M5Vo7LhMsMFb1hPtY68kbDUqD01RuMD5KhGIngCRRBZJriRQclUCJS4D3jr/Frw9ruNGh+NTIvIwdv0Y2brU= opensource@scaleway.com"
	sshKey2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQClWb6lwORneAv3JvuMXkQmnt3g6vJ6fHjcojZfOrSdKfx6BOgE6BauAw1wibD0eLYVzZtC9onIjSeDETig0psSILv2O7I57XlaWdqTDjw7w74opVCsx8X1VelU3NwINkXEy2wR2htgJ2zUASniJw+Wt644WkSICcMPnSw9boE1vWwSsykloJr3MdXKn2pwVZ26lumoutHlm/XQ8DWGHh6gxE8DQIVKrCqVFp2ewPsCH4I/LtcWBAX9NTuUhG0uotunhYUoajKN7N0W3XeOxobjh2kvWZlQPjHEcUdKroUBGpIeEAu0i89lFgwqvk0Zdlc+FkJwaSliv3cikXToLK+H20BDlb0hwS/zkjpHKvNenjP7M+awE4aapOtZNvO1fsLfzeqaWsYvAOesq4bUfL6UkfVSZ/Z+pU7U3DsUviPhzLJ4Sf6daQYwgdCDXLfcsJprDZ+g0k0+DeZkJQv5/9dg3yLf537eXoJ/H1kNmQroCBE6Rra6MNQPMtCBrD8I8kE= user@user-ThinkPad-T14s-Gen-6"

	// For now, we need to explicitly define the project_id for the CI's Acceptance Tests to pass.
	// We should find a more appropriate solution in the future.
	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		projectID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			iamchecks.CheckSSHKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key1" {
						name       = "tf-test-acc-instance-tmpl-admin-ssh-key1"
						public_key = %q
					}

					resource "scaleway_instance_template" "main" {
						project_id  = %q
						name        = "tf-test-acc-instance-tmpl-admin-ssh-key"
						tags        = [ "terraform-test", "scaleway_instance_template", "windows_rdp_ssh_key" ]
						zone        = "fr-par-1"
						server_type = "POP2-4C-16G-WIN"

						windows_rdp_ssh_key_id = scaleway_iam_ssh_key.key1.id
					}`, sshKey1, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					iamchecks.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.key1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-4C-16G-WIN"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "windows_rdp_ssh_key_id", "scaleway_iam_ssh_key.key1", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
			},
			{
				ResourceName:      "scaleway_instance_template.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key1" {
						name       = "tf-test-acc-instance-tmpl-admin-ssh-key1"
						public_key = %q
					}

					resource "scaleway_iam_ssh_key" "key2" {
						name       = "tf-test-acc-instance-tmpl-admin-ssh-key2"
						public_key = %q
					}

					resource "scaleway_instance_template" "main" {
						project_id  = %q
						name        = "tf-test-acc-instance-tmpl-admin-ssh-key"
						tags        = [ "terraform-test", "scaleway_instance_template", "windows_rdp_ssh_key" ]
						zone        = "fr-par-1"
						server_type = "POP2-4C-16G-WIN"

						windows_rdp_ssh_key_id = scaleway_iam_ssh_key.key2.id
					}`, sshKey1, sshKey2, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					iamchecks.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.key2"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-4C-16G-WIN"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_template.main", "windows_rdp_ssh_key_id", "scaleway_iam_ssh_key.key2", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_template.main", &templateID),
				),
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
						tags   = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_vpc_private_network.pn0.project_id
						name        = "tf-test-acc-instance-tmpl-private-networks"
						tags        = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone        =  "nl-ams-1"
						server_type = "PLAY2-MICRO"

						private_networks = [
							scaleway_vpc_private_network.pn0.id,
							substr(scaleway_vpc_private_network.pn1.id, 7, -1),
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "PLAY2-MICRO"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "2"),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "private_networks", "scaleway_vpc_private_network.pn0", "id", false),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "private_networks", "scaleway_vpc_private_network.pn1", "id", true),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc" {
						region = "nl-ams"
					}

					resource "scaleway_vpc_private_network" "pn0" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_vpc_private_network.pn0.project_id
						name        = "tf-test-acc-instance-tmpl-private-networks"
						tags        = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone        =  "nl-ams-1"
						server_type = "PLAY2-MICRO"

						private_networks = [
							scaleway_vpc_private_network.pn1.id,
							scaleway_vpc_private_network.pn2.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "private_networks.#", "2"),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "private_networks", "scaleway_vpc_private_network.pn1", "id", false),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "private_networks", "scaleway_vpc_private_network.pn2", "id", false),
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
						tags   = [ "scaleway_instance_template", "private-networks", "pn0" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn1" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn1" ]
						region = "nl-ams"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						vpc_id = scaleway_vpc.vpc.id
						tags   = [ "scaleway_instance_template", "private-networks", "pn2" ]
						region = "nl-ams"
					}

					resource "scaleway_instance_template" "main" {
						project_id  = scaleway_vpc_private_network.pn0.project_id
						name        = "tf-test-acc-instance-tmpl-private-networks"
						tags        = [ "terraform-test", "scaleway_instance_template", "private-networks" ]
						zone        =  "nl-ams-1"
						server_type = "PLAY2-MICRO"

						private_networks = []
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
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

func TestAccInstanceTemplateResource_Filesystems(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTemplateDestroyed(tt),
			filetestfuncs.CheckFileDestroy(tt),
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
					}

					resource "scaleway_instance_template" "main" {
						project_id        = scaleway_file_filesystem.fs0.project_id
						name              = "tf-test-acc-instance-tmpl-filesystems"
						tags              = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type       = "POP2-8C-32G"
						public_ipv4_count = 1
						public_ipv6_count = 2

						filesystem_ids = [
							scaleway_file_filesystem.fs0.id,
							scaleway_file_filesystem.fs1.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-8C-32G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "2"),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "filesystem_ids", "scaleway_file_filesystem.fs0", "id", false),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "filesystem_ids", "scaleway_file_filesystem.fs1", "id", false),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv4_count", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv6_count", "2"),
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
						project_id  = scaleway_file_filesystem.fs0.project_id
						name        = "tf-test-acc-instance-tmpl-filesystems"
						tags        = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type = "POP2-8C-32G"

						filesystem_ids = [
							substr(scaleway_file_filesystem.fs1.id, 7, -1),
							scaleway_file_filesystem.fs2.id,
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "server_type", "POP2-8C-32G"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "filesystem_ids.#", "2"),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "filesystem_ids", "scaleway_file_filesystem.fs1", "id", true),
					acctest.CheckResourceAttrPairInSet("scaleway_instance_template.main", "filesystem_ids", "scaleway_file_filesystem.fs2", "id", false),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv4_count", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_template.main", "public_ipv6_count", "0"),
				),
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
						project_id  = scaleway_file_filesystem.fs0.project_id
						name        = "tf-test-acc-instance-tmpl-filesystems"
						tags        = [ "terraform-test", "scaleway_instance_template", "filesystems" ]
						server_type = "POP2-8C-32G"

						filesystem_ids = []
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isTemplatePresent(tt, "scaleway_instance_template.main"),
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
