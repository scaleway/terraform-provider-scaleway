package instance_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	"github.com/stretchr/testify/assert"
)

func TestAccServer_Migrate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := new("")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-migrate"
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						tags  = [ "terraform-test", "scaleway_instance_server", "migrate" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PRO2-XXS"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.main", serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-migrate"
						image = "ubuntu_jammy"
						type  = "PRO2-XS"
						tags  = [ "terraform-test", "scaleway_instance_server", "migrate" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PRO2-XS"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.main", serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-migrate"
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						tags  = [ "terraform-test", "scaleway_instance_server", "migrate" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PRO2-XXS"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.main", serverID),
				),
			},
		},
	})
}

func TestAccServer_Migrate_ReplaceOnTypeChange(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := new("")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// DEV1-M
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-replace-on-type-change"
					  image = "ubuntu_jammy"
					  type  = "DEV1-M"

					  tags = [ "terraform-test", "scaleway_instance_server", "replace_on_type_change" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-replace-on-type-change"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", serverID),
				),
			},
			{
				// DEV1-S
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-replace-on-type-change"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  replace_on_type_change  = true

					  tags = [ "terraform-test", "scaleway_instance_server", "replace_on_type_change" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-replace-on-type-change"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					acctest.CheckResourceIDChanged("scaleway_instance_server.base", serverID),
				),
			},
		},
	})
}

func TestAccServer_Migrate_InvalidLocalVolumeSize(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-migrate-invalid-local-volume-size"
						image = "ubuntu_jammy"
						type  = "DEV1-L"
						tags  = [ "terraform-test", "scaleway_instance_server", "migrate_invalid_local_volume_size" ]
						root_volume {
							volume_type = "l_ssd"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "DEV1-L"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", "l_ssd"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-migrate-invalid-local-volume-size"
						image = "ubuntu_jammy"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "migrate_invalid_local_volume_size" ]
						root_volume {
							volume_type = "l_ssd"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", "l_ssd"),
				),
				ExpectError: regexp.MustCompile("cannot change server type"),
				PlanOnly:    true,
			},
		},
	})
}

func TestGetEndOfServiceDate(t *testing.T) {
	tt := acctest.NewTestTools(t)

	client := meta.ExtractScwClient(tt.Meta)
	defer tt.Cleanup()

	eosDate, err := instance.GetEndOfServiceDate(t.Context(), client, "fr-par-1", "ENT1-S")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "2025-09-01", eosDate)
}
