package partner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	partnerSDK "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var DestroyWaitTimeout = 3 * time.Minute

func TestAccPartnerOrganizationResource_Basic(t *testing.T) {
	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := uuid.NewString() + "@test.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email           = "%s"
						organization_name = "tf_test_partner_org"
						partner_id = "%s"
						owner_firstname = "John"
						owner_lastname  = "Doe"
						customer_id     = "customer-123"
					}
				`, email, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "email", email),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "organization_name", "tf_test_partner_org"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "partner_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "id"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "owner_firstname", "John"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "owner_lastname", "Doe"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "customer_id", "customer-123"),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "locked_by"),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "locked_at"),
				),
			},
			{
				ResourceName:            "scaleway_partner_organization.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"partner_id"},
			},
		},
	})
}

func TestAccPartnerOrganizationResource_Update(t *testing.T) {
	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := uuid.NewString() + "@test.test"
	updatedEmail := uuid.NewString() + "@test.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email           = "%s"
						organization_name = "tf_test_partner_org"
						partner_id = "%s"
						owner_firstname = "John"
						owner_lastname  = "Doe"
						customer_id     = "customer-123"
					}
				`, email, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "email", email),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "organization_name", "tf_test_partner_org"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email           = "%s"
						organization_name = "tf_test_partner_org_updated"
						partner_id = "%s"
						owner_firstname = "Jane"
						owner_lastname  = "Smith"
						customer_id     = "customer-456"
						comment         = "Updated organization"
					}
				`, updatedEmail, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "email", updatedEmail),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "organization_name", "tf_test_partner_org_updated"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "owner_firstname", "Jane"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "owner_lastname", "Smith"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "customer_id", "customer-456"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "comment", "Updated organization"),
				),
			},
		},
	})
}

func TestAccPartnerOrganizationResource_WithPhoneNumber(t *testing.T) {
	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := uuid.NewString() + "@test.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email           = "%s"
						organization_name = "tf_test_partner_org_phone"
						partner_id = "%s"
						owner_firstname = "John"
						owner_lastname  = "Doe"
						customer_id     = "customer-789"
						phone_number    = "+33123456789"
						siren_number    = "123456782"
					}
				`, email, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "email", email),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "phone_number", "+33123456789"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "siren_number", "123456782"),
				),
			},
			{
				RefreshState: true,
			},
			{
				ResourceName:            "scaleway_partner_organization.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"partner_id"},
			},
		},
	})
}

func TestAccPartnerOrganizationResource_DefaultPartnerID(t *testing.T) {
	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := uuid.NewString() + "@test.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email           = "%s"
						organization_name = "tf_test_partner_org_default"
						owner_firstname = "John"
						owner_lastname  = "Doe"
						customer_id     = "customer-default"
					}
				`, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "partner_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "id"),
					resource.TestCheckResourceAttr("scaleway_partner_organization.main", "customer_id", "customer-default"),
				),
			},
		},
	})
}

func IsPartnerOrganizationLocked(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_partner_organization" {
				continue
			}

			partnerAPI := partnerSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

			err := retry.RetryContext(context.Background(), DestroyWaitTimeout, func() *retry.RetryError {
				_, err := partnerAPI.LockOrganization(&partnerSDK.LockOrganizationRequest{
					OrganizationID: rs.Primary.ID,
				})
				if err != nil {
					if httperrors.Is404(err) {
						return retry.RetryableError(errors.New("partner organization not yet retrievable"))
					}

					return retry.NonRetryableError(err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
