package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccServer_DefaultTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	// Inject default tags into the provider Meta for testing.
	tt.Meta.SetDefaultTags([]string{"env:dev", "team:scaleway"})

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-default-tags"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					  tags  = [ "terraform-test", "server" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					// tags should only contain resource-specific tags (no defaults)
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "server"),
					// tags_all should contain both default and resource-specific tags
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.#", "4"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.0", "env:dev"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.1", "team:scaleway"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.2", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.3", "server"),
				),
			},
			{
				ResourceName:            "scaleway_instance_server.base",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image"},
			},
		},
	})
}

func TestAccServer_DefaultTags_NoResourceTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tt.Meta.SetDefaultTags([]string{"env:dev"})

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-default-tags-none"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					// tags should be empty (no resource-specific tags)
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "0"),
					// tags_all should contain only default tags
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags_all.0", "env:dev"),
				),
			},
		},
	})
}
