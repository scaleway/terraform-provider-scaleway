package documentdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	documentdbtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/documentdb/testfuncs"
)

func init() {
	documentdbtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
