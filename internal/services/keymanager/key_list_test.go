package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListKeyManagerKeys_ByProjectIDs(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByProjectIDs because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-km-by-proj-id-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "by_project" {
						provider = scaleway

						config {
							project_ids = [scaleway_key_manager_key.key1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_key_manager_key.by_project", 1),
				},
			},
		},
	})
}

func TestAccListKeyManagerKeys_ByName(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByName because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-km-by-name-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Config: `
					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-km-by-name-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}

					resource "scaleway_key_manager_key" "key2" {
						name        = "tf-test-km-by-name-2"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "by_name" {
						provider = scaleway

						config {
							project_ids = [scaleway_key_manager_key.key1.project_id]
							name        = "tf-test-km-by-name-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_key_manager_key.by_name", 1),
				},
			},
		},
	})
}

func TestAccListKeyManagerKeys_ByUsage(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByUsage because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_key_manager_key" "symmetric_key" {
						name        = "tf-test-km-by-usage-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Config: `

					resource "scaleway_key_manager_key" "symmetric_key" {
						name        = "tf-test-km-by-usage-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}

					resource "scaleway_key_manager_key" "asymmetric_key" {
						name        = "tf-test-km-by-usage-2"
						usage        = "asymmetric_encryption"
						algorithm    = "rsa_oaep_4096_sha256"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "by_usage" {
						provider = scaleway

						config {
							project_ids = [scaleway_key_manager_key.asymmetric_key.project_id]
							usage        = "asymmetric_encryption"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_key_manager_key.by_usage", 1),
				},
			},
		},
	})
}

func TestAccListKeyManagerKeys_ByTags(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByTags because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_key_manager_key" "key_with_tags" {
						name        = "tf-test-km-by-tags-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						tags        = ["env:test", "team:test"]
						unprotected = true
					}
				`,
			},
			{
				Config: `
					resource "scaleway_key_manager_key" "key_with_tags" {
						name        = "tf-test-km-by-tags-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						tags        = ["env:test", "team:test"]
						unprotected = true
					}

					resource "scaleway_key_manager_key" "key_without_tags" {
						name        = "tf-test-km-by-tags-2"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "by_tags" {
						provider = scaleway

						config {
							project_ids = [scaleway_key_manager_key.key_with_tags.project_id]
							tags        = ["env:test"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_key_manager_key.by_tags", 1),
				},
			},
		},
	})
}

func TestAccListKeyManagerKeys_ByScheduledForDeletion(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByScheduledForDeletion because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			accounttestfuncs.IsProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `

					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-km-by-scheduled-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Config: `

					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-km-by-scheduled-1"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}

					resource "scaleway_key_manager_key" "key2" {
						name        = "tf-test-km-by-scheduled-2"
						usage        = "symmetric_encryption"
						algorithm    = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "not_scheduled" {
						provider = scaleway

						config {
							project_ids           = [scaleway_key_manager_key.key1.project_id]
							scheduled_for_deletion = false
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_key_manager_key.not_scheduled", 2),
				},
			},
		},
	})
}

func TestAccListKeyManagerKeys_ByRegions(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListKeyManagerKeys_ByRegions because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_key_manager_key" "key1" {
						name        = "tf-test-key-by-region"
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						unprotected = true
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_key_manager_key" "by_region" {
						provider = scaleway

						config {
							project_ids = [scaleway_key_manager_key.key1.project_id]
							regions 	= [scaleway_key_manager_key.key1.region]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_key_manager_key.by_region", 1),
				},
			},
		},
	})
}
