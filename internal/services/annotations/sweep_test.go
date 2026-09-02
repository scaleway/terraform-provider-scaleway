package annotations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	annotationstestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations/testfuncs"
)

func init() {
	annotationstestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}