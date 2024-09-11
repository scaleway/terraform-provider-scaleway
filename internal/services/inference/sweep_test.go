package inference_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
	"testing"
)

func init() {
	inferencetestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
