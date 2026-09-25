package acctest

import (
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

const (
	// ProviderNameEcho is the provider name for the echo provider used for testing ephemeral resources
	ProviderNameEcho = "echo"
)

// ProtoV6ProviderFactoriesEcho returns a map of provider factories for the echo provider.
// This is used for testing ephemeral resources by echoing their data back for verification.
func ProtoV6ProviderFactoriesEcho() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		ProviderNameEcho: echoprovider.NewProviderServer(),
	}
}

// MergeProviderFactories merges two provider factory maps together.
// This is useful when combining the test tools provider factories with additional providers like echo.
func MergeProviderFactories(base map[string]func() (tfprotov6.ProviderServer, error), additional map[string]func() (tfprotov6.ProviderServer, error)) map[string]func() (tfprotov6.ProviderServer, error) {
	result := make(map[string]func() (tfprotov6.ProviderServer, error), len(base)+len(additional))
	maps.Copy(result, base)

	maps.Copy(result, additional)

	return result
}

// ConfigWithEchoProvider returns a Terraform configuration snippet that sets up the echo provider.
// The echo provider echoes back the data from ephemeral resources for verification purposes.
//
// ephemeralResourceData should be a reference to the ephemeral resource's data attribute, e.g.:
//   - ephemeral.scaleway_iam_api_key.main
//
// Example usage:
//
//	ConfigCompose(
//		ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"),
//		`resource "echo" "test" {}`,
//	)
func ConfigWithEchoProvider(ephemeralResourceData string) string {
	// lintignore:AT004
	return fmt.Sprintf(`
provider "echo" {
  data = %s
}

resource "echo" "test" {}
`, ephemeralResourceData)
}
