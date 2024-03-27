package scaleway

import (
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func AccountV3ProjectAPI(m interface{}) *accountV3.ProjectAPI {
	return accountV3.NewProjectAPI(meta.ExtractScwClient(m))
}
