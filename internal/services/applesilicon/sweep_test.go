package applesilicon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	applesilicontestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/applesilicon/testfuncs"
)

func init() {
	applesilicontestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
