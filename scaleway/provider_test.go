package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
}

func testAccPreCheck(_ *testing.T) {}
