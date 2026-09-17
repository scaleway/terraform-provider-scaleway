package secret_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccListSecrets_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecrets_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-list-1"
						path = "/"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-list-1"
						path = "/"
					}

					resource "scaleway_secret" "secret2" {
						name = "test-secret-list-2"
						path = "/"
						tags = ["test-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret" "all" {
						provider = scaleway

						config {
							project_ids = [scaleway_secret.secret1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_secret.all", 2),
				},
			},
		},
	})
}

func TestAccListSecrets_ByName(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecrets_ByName because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-filter-1"
						path = "/"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-filter-1"
						path = "/"
					}

					resource "scaleway_secret" "secret2" {
						name = "test-secret-filter-2"
						path = "/"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret" "by_name" {
						provider = scaleway

						config {
							project_ids = [scaleway_secret.secret1.project_id]
							name = "test-secret-filter-1"
							
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_secret.by_name", 1),
				},
			},
		},
	})
}

func TestAccListSecrets_ByTags(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecrets_ByTags because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-tag-1"
						path = "/"
						tags = ["production"]
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-tag-1"
						path = "/"
						tags = ["production"]
					}

					resource "scaleway_secret" "secret2" {
						name = "test-secret-tag-2"
						path = "/"
						tags = ["development"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret" "by_tags" {
						provider = scaleway

						config {
							project_ids = [scaleway_secret.secret1.project_id]
							tags = ["production"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_secret.by_tags", 1),
				},
			},
		},
	})
}
