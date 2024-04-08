package rdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	rdbtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func init() {
	rdbtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
