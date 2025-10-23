package datawarehouse_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	datawarehousetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/datawarehouse/testfuncs"
)

func init() {
	datawarehousetestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
