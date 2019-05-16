package scaleway

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	homedir "github.com/mitchellh/go-homedir"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"scaleway": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SCALEWAY_ORGANIZATION"); v == "" {
		if path, err := homedir.Expand("~/.scwrc"); err == nil {
			scwAPIKey, scwOrganization, err := readDeprecatedScalewayConfig(path)
			if err != nil {
				t.Fatalf("failed falling back to %s: %v", path, err)
			}
			if scwAPIKey == "" && scwOrganization == "" {
				t.Fatal("SCALEWAY_TOKEN must be set for acceptance tests")
			}
			return
		}
		t.Fatal("SCALEWAY_ORGANIZATION must be set for acceptance tests")
	}
	tokenFromAccessKey := os.Getenv("SCALEWAY_ACCESS_KEY")
	token := os.Getenv("SCALEWAY_TOKEN")
	if token == "" && tokenFromAccessKey == "" {
		t.Fatal("SCALEWAY_TOKEN must be set for acceptance tests")
	}
}
