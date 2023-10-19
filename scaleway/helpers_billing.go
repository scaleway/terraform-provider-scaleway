package scaleway

import (
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2alpha1"
)

// billingAPI returns a new billing API.
func billingAPI(m interface{}) *billing.API {
	meta := m.(*Meta)
	return billing.NewAPI(meta.scwClient)
}
