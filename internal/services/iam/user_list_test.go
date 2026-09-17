package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	identitycheck "github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest/identity"
	iamtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

func TestAccListIAMUsers_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMUsers_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamtestfuncs.CheckUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "user1" {
						email = "test-user-list-1@example.com"
						username = "test-user-list-1"
						tags = ["toto"]
					}
				`,
			},
			{
				Config: `
					resource "scaleway_iam_user" "user1" {
						email = "test-user-list-1@example.com"
						username = "test-user-list-1"
						tags = ["toto"]
					}

					resource "scaleway_iam_user" "user2" {
						email = "test-user-list-2@example.com"
						username = "test-user-list-2"
						tags = ["test-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_user" "all" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_user.user2.organization_id
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_user.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_user" "by_tag" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_user.user2.organization_id
							tag              = "test-tag"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_user.by_tag", 1),
				},
			},
		},
	})
}

func TestAccListIAMUsers_MFA(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMUsers_MFA because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "scaleway_iam_user.user1"
	identity := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamtestfuncs.CheckUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "user1" {
						email = "test-user-mfa@example.com"
						username = "test-user-mfa"
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("username"), knownvalue.StringExact("test-user-mfa")),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_user" "by_mfa_false" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_user.user1.organization_id
							mfa              = false
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					identitycheck.ExpectIdentityFunc("scaleway_iam_user.by_mfa_false", identity.Checks()),
					querycheck.ExpectResourceDisplayName("scaleway_iam_user.by_mfa_false", identitycheck.FilterByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact("test-user-mfa@example.com")),
					identitycheck.ExpectNoResourceObject("scaleway_iam_user.by_mfa_false", identitycheck.FilterByResourceIdentityFunc(identity.Checks())),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_user" "by_mfa_true" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_user.user1.organization_id
							mfa              = true
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					identitycheck.ExpectNoIdentityFunc("scaleway_iam_user.by_mfa_true", identity.Checks()),
				},
			},
		},
	})
}
