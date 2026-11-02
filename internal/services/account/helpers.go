package account

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const defaultRetryInterval = 5 * time.Second

func NewProjectAPI(m any) *accountSDK.ProjectAPI {
	return accountSDK.NewProjectAPI(meta.ExtractScwClient(m))
}

func GetOrganizationID(m any, d *schema.ResourceData) *string {
	orgID, orgIDExist := d.GetOk("organization_id")

	if orgIDExist {
		return types.ExpandStringPtr(orgID)
	}

	defaultOrgID, defaultOrgIDExists := meta.ExtractScwClient(m).GetDefaultOrganizationID()
	if defaultOrgIDExists {
		return types.ExpandStringPtr(defaultOrgID)
	}

	return nil
}

func isProjectNotUsableError(err error) bool {
	if err == nil {
		return false
	}

	if responseError, ok := errors.AsType[*scw.PreconditionFailedError](err); ok && responseError.Precondition == "resource_not_usable" {
		return true
	}

	return false
}

func isProjectStillInUseError(err error) bool {
	if err == nil {
		return false
	}

	if responseError, ok := errors.AsType[*scw.PreconditionFailedError](err); ok && responseError.Precondition == "resource_still_in_use" {
		return true
	}

	return false
}

func waitForProjectDeleted(ctx context.Context, api *accountSDK.ProjectAPI, projectID string, timeout time.Duration) error {
	retryInterval := defaultRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	_, err := api.WaitForProject(&accountSDK.WaitForProjectRequest{
		ProjectID:     projectID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && httperrors.Is404(err) {
		return nil
	}

	return err
}
