package account

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func NewProjectAPI(m interface{}) *accountSDK.ProjectAPI {
	return accountSDK.NewProjectAPI(meta.ExtractScwClient(m))
}

func GetOrganizationID(m interface{}, d *schema.ResourceData) *string {
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

	responseError := &scw.PreconditionFailedError{}
	if errors.As(err, &responseError) && responseError.Precondition == "resource_not_usable" {
		return true
	}

	return false
}
