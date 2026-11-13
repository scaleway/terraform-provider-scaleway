package billing_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	billingtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/billing/testfuncs"
)

func init() {
	billingtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
