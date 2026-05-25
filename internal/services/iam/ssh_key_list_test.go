package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

func TestAccListIAMSSHKeys_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMSSHKeys_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamchecks.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key1" {
					  name       = "test-ssh-key-list-1"
					  public_key = "%[1]s"
					}
				`, SSHKey),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key1" {
					  name       = "test-ssh-key-list-1"
					  public_key = "%[1]s"
					}

					resource "scaleway_iam_ssh_key" "key2" {
					  name       = "test-ssh-key-list-2"
					  public_key = "%[2]s"
					}
				`, SSHKey, SSHKey+"-2"),
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_ssh_key" "by_name" {
					  provider = scaleway

					  config {
						project_ids = [scaleway_iam_ssh_key.key1.project_id]
						name        = "test-ssh-key-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_ssh_key.by_name", 1),
				},
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "key1" {
					  name       = "test-ssh-key-list-1"
					  public_key = "%[1]s"
					}

					resource "scaleway_iam_ssh_key" "key2" {
					  name       = "test-ssh-key-list-2"
					  public_key = "%[2]s"
					  disabled   = true
					}
				`, SSHKey, SSHKey+"-2"),
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_ssh_key" "by_disabled" {
					  provider = scaleway

					  config {
						project_ids = [scaleway_iam_ssh_key.key2.project_id]
						disabled    = true
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_ssh_key.by_disabled", 1),
				},
			},
		},
	})
}
