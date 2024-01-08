package scaleway

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

func init() {
	resource.AddTestSweepers("scaleway_rdb_log_policy", &resource.Sweeper{
		Name: "scaleway_rdb_log_policy",
		//Dependencies: nil,
		F: testSweepRDBLogPolicy,
	})
}
