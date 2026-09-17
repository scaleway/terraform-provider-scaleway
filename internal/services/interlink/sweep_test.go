package interlink_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	interlinktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/interlink/testfuncs"
)

func init() {
	interlinktestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
