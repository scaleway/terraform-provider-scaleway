package scaleway

import (
	"time"

	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
)

const (
	defaultAccountSSHKeyTimeout = 1 * time.Minute
)

// accountAPI returns a new account API.
func accountAPI(m interface{}) *account.API {
	meta := m.(*Meta)
	return account.NewAPI(meta.scwClient)
}

// accountV2API supports projects.
func accountV2API(m interface{}) *accountV2.API {
	meta := m.(*Meta)
	return accountV2.NewAPI(meta.scwClient)
}
