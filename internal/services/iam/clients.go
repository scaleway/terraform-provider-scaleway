package iam

import (
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// NewAPI returns a new iam API for a Create request
func NewAPI(m any) *iam.API {
	return iam.NewAPI(meta.ExtractScwClient(m))
}
