package testfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// AddTestSweepers adds datawarehouse sweepers
func AddTestSweepers() {
	// For now, no specific sweepers needed for datawarehouse
	// This function exists to match the pattern used by other services
	_ = resource.TestMain
}
