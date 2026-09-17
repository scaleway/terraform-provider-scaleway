package secret_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccListSecretVersions_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecretVersions_Basic because list resources are not yet supported on OpenTofu")
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
						name = "test-secret-version-list-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-version-list-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}

					resource "scaleway_secret_version" "version2" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret_version" "all" {
						provider = scaleway

						config {
							secret_ids = [scaleway_secret.secret1.id]
							project_ids = [scaleway_secret.secret1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_secret_version.all", 2),
				},
			},
		},
	})
}

func TestAccListSecretVersions_ByStatus(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecretVersions_ByStatus because list resources are not yet supported on OpenTofu")
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
						name = "test-secret-version-filter-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-version-filter-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}

					resource "scaleway_secret_version" "version2" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-2"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-version-filter-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}

					resource "scaleway_secret_version" "version2" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret_version" "enabled" {
						provider = scaleway

						config {
							secret_ids = [scaleway_secret.secret1.id]
							status = ["enabled"]
							project_ids = [scaleway_secret.secret1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_secret_version.enabled", 2),
				},
			},
		},
	})
}

func TestAccListSecretVersions_AllSecrets(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecretVersions_AllSecrets because list resources are not yet supported on OpenTofu")
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
						name = "test-secret-version-all-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-version-all-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}

					resource "scaleway_secret" "secret2" {
						name = "test-secret-version-all-2"
						path = "/"
					}

					resource "scaleway_secret_version" "version2" {
						secret_id = scaleway_secret.secret2.id
						data = "my-secret-data-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret_version" "all_secrets" {
						provider = scaleway

						config {
							secret_ids = ["*"]
							project_ids = [scaleway_secret.secret1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_secret_version.all_secrets", 2),
				},
			},
		},
	})
}

func TestAccListSecretVersions_MultipleSecrets(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListSecretVersions_MultipleSecrets because list resources are not yet supported on OpenTofu")
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
						name = "test-secret-multi-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_secret" "secret1" {
						name = "test-secret-multi-1"
						path = "/"
					}

					resource "scaleway_secret_version" "version1" {
						secret_id = scaleway_secret.secret1.id
						data = "my-secret-data-1"
					}

					resource "scaleway_secret" "secret2" {
						name = "test-secret-multi-2"
						path = "/"
					}

					resource "scaleway_secret_version" "version2" {
						secret_id = scaleway_secret.secret2.id
						data = "my-secret-data-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_secret_version" "multi" {
						provider = scaleway

						config {
							secret_ids = [scaleway_secret.secret1.id, scaleway_secret.secret2.id]
							project_ids = [scaleway_secret.secret1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_secret_version.multi", 2),
				},
			},
		},
	})
}
