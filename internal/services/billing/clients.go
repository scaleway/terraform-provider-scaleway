package billing

import (
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// billingAPI returns a new billing API.
func billingAPI(m any) *billing.API {
	return billing.NewAPI(meta.ExtractScwClient(m))
}
