package opensearch_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	opensearchtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch/testfuncs"
)

func init() {
	opensearchtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
