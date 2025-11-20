package audittrail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/account/v3"
	audit_trail "github.com/scaleway/scaleway-sdk-go/api/audit_trail/v1alpha1"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

const (
	defaultAuditTrailEventsTimeout = 20 * time.Second
	destroyWaitTimeout             = 3 * time.Minute
	dummyID                        = "AB7BD9BF-E1BD-41E8-9F1D-F16A2E3F3925"
	serviceName                    = "scaleway.secret_manager.v1beta1.Api"
	productName                    = "secret-manager"
	methodCreate                   = "CreateSecret"
	resourceType                   = "secret_manager_secret"
	resourceName                   = "scaleway_secret.main"
	secretName                     = "scalewayDataSourceAuditTrailSecret"
)

func getAuditEvents(ctx context.Context, api *audit_trail.API, projectID string, method string) (*audit_trail.ListEventsResponse, error) {
	resp, err := api.ListEvents(&audit_trail.ListEventsRequest{
		ProjectID:  &projectID,
		MethodName: &method,
	}, scw.WithContext(ctx))

	return resp, err
}

func waitForAuditTrailEvents(t *testing.T, ctx context.Context, api *audit_trail.API, project *account.Project) {
	t.Helper()

	// Retry call with exponential backoff.
	// Wait up to 20 seconds for audit event to appear.
	err := retry.RetryContext(ctx, defaultAuditTrailEventsTimeout, func() *retry.RetryError {
		events, err := getAuditEvents(ctx, api, project.ID, methodCreate)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to list audit events: %w", err))
		}

		if len(events.Events) > 0 {
			return nil // Success
		}

		// Not found yet
		return retry.RetryableError(errors.New("audit event not found yet for resource, retrying..."))
	})
	if err != nil {
		t.Fatalf("timed out waiting for audit trail event: %v", err)
	}
}

// Bundle of checks for audit trail event data source testing
func createEventDataSourceChecks(eventsDataSourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(
			eventsDataSourceName, "events.0.locality",
			resourceName, "region",
		),
		resource.TestCheckResourceAttrSet(eventsDataSourceName, "events.0.principal_id"),
		resource.TestCheckResourceAttrSet(eventsDataSourceName, "events.0.source_ip"),
		resource.TestCheckResourceAttrSet(eventsDataSourceName, "events.0.user_agent"),
		resource.TestCheckResourceAttr(
			eventsDataSourceName, "events.0.product_name", productName,
		),
		resource.TestCheckResourceAttr(
			eventsDataSourceName, "events.0.service_name", serviceName,
		),
		resource.TestCheckResourceAttr(
			eventsDataSourceName, "events.0.method_name", methodCreate,
		),
		testCheckSecretIDEqualsEventResourceID(eventsDataSourceName, resourceName),
		resource.TestCheckResourceAttrPair(
			eventsDataSourceName, "events.0.resources.0.name",
			resourceName, "name",
		),
		resource.TestCheckResourceAttr(
			eventsDataSourceName, "events.0.resources.0.type", resourceType,
		),
		resource.TestCheckResourceAttrSet(eventsDataSourceName, "events.0.request_id"),
		resource.TestCheckResourceAttrSet(eventsDataSourceName, "events.0.request_body"),
		resource.TestCheckResourceAttr(
			eventsDataSourceName, "events.0.status_code", "200",
		),
	)
}

// testAccCheckSecretIDEqualsEventResourceID is a convenience wrapper for checking
// secret resource IDs against audit trail event resource IDs.
func testCheckSecretIDEqualsEventResourceID(eventsDataSourceName, resourceName string) resource.TestCheckFunc {
	return checkSecretResourceIDMatchesEvent(
		resourceName,
		"id",
		eventsDataSourceName,
		"events.0.resources.0.id",
		"events.0.locality",
	)
}

// scaleway_secret id is formatted as ${scaleway_secret.main.region}/${scaleway_secret.main.id}
// this function enables checking the event region and id against the scaleway_secret formatted id.
func checkSecretResourceIDMatchesEvent(resourceName, resourceIDAttr, eventsDataSourceName, eventResourceIDAttr, eventLocalityAttr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Retrieve the secret resource and relevant attribute
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		resourceID, ok := resource.Primary.Attributes[resourceIDAttr]
		if !ok {
			return fmt.Errorf("attribute not found: %s %s", resource, resourceIDAttr)
		}

		// Retrieve the events data source and relevant attributes
		events, ok := s.RootModule().Resources[eventsDataSourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", eventsDataSourceName)
		}

		eventResourceID, ok := events.Primary.Attributes[eventResourceIDAttr]
		if !ok {
			return fmt.Errorf("attribute not found: %s %s", events, eventResourceIDAttr)
		}

		eventLocality, ok := events.Primary.Attributes[eventLocalityAttr]
		if !ok {
			return fmt.Errorf("attribute not found: %s %s", events, eventLocalityAttr)
		}

		// Format event resource ID to match the secret resource ID pattern
		eventResourceFormattedID := fmt.Sprintf("%s/%s", eventLocality, eventResourceID)

		if resourceID != eventResourceFormattedID {
			return fmt.Errorf("expected %s, got %s", resourceID, eventResourceFormattedID)
		}

		return nil
	}
}

func testAccCheckSecretDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, destroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_secret" {
					continue
				}

				api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				sec, err := api.GetSecret(&secretSDK.GetSecretRequest{
					SecretID: id,
					Region:   region,
				})

				switch {
				case err == nil && sec != nil && sec.DeletionRequestedAt != nil:
					// Soft-deleted (scheduled for deletion), treat as destroyed for tests
					continue
				case httperrors.Is404(err):
					continue
				case err != nil:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
