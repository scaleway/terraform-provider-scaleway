package scaleway

import iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"

// instanceAPIWithZone returns a new iam API for a Create request
func iamAPI(m interface{}) *iam.API {
	meta := m.(*Meta)
	return iam.NewAPI(meta.scwClient)
}
