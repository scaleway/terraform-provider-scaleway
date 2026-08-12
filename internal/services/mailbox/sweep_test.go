package mailbox_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func init() {
	mailboxtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
