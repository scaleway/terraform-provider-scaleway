package domain_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	domaintestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain/testfuncs"
	temtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem/testfuncs"
)

func init() {
	// Register TEM first so scaleway_domain_zone Dependencies can revoke
	// before deleting DNS zones (same pattern as vpc + ipam sweepers).
	temtestfuncs.AddTestSweepers()
	domaintestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
