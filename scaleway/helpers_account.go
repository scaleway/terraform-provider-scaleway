package scaleway

import (
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
)

func accountV3ProjectAPI(m interface{}) *accountV3.ProjectAPI {
	meta := m.(*Meta)
	return accountV3.NewProjectAPI(meta.scwClient)
}
