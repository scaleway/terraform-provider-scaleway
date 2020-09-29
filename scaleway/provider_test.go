package scaleway

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	p := Provider()()
	testAccProvider = p
	version += "-tftest"
	testAccProviders = map[string]*schema.Provider{
		"scaleway": p,
	}

	old := testAccProvider.ConfigureFunc
	testAccProvider.ConfigureFunc = func(data *schema.ResourceData) (i interface{}, e error) {
		_ = data.Set("region", "fr-par")
		_ = data.Set("zone", "fr-par-1")
		return old(data)
	}

}

func TestProvider(t *testing.T) {
	p := Provider()()
	if err := p.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {

	// Handle new config system first
	_, _ = scw.MigrateLegacyConfig()
	config, err := scw.LoadConfig()
	if err == nil {
		activeProfile, err := config.GetActiveProfile()
		if err == nil {
			if activeProfile.AccessKey != nil && activeProfile.SecretKey != nil {
				return
			}
		}
	}
	envProfile := scw.LoadEnvProfile()
	if envProfile.AccessKey != nil && envProfile.SecretKey != nil {
		return
	}

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
