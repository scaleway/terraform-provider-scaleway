package messageq_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	messageqtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/messageq/testfuncs"
)

func init() {
	messageqtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
