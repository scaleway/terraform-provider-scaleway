package scaleway

import (
	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
)

// accountV2API supports projects.
func accountV2API(m interface{}) *accountV2.API {
	meta := m.(*Meta)
	return accountV2.NewAPI(meta.scwClient)
}
