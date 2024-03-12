package scaleway

import (
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func accountV3ProjectAPI(m interface{}) *accountV3.ProjectAPI {
	meta := m.(*meta.Meta)
	return accountV3.NewProjectAPI(meta.ScwClient())
}
