package scaleway

import (
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
)

// accountAPI returns a new account API.
func accountAPI(m interface{}) *account.API {
	meta := m.(*Meta)

	return account.NewAPI(meta.scwClient)
}
