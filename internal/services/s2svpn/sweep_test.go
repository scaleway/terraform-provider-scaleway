package s2svpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	s2svpntestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/s2svpn/testfuncs"
)

func init() {
	s2svpntestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
