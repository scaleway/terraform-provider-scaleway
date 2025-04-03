package baremetal_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const (
	offerISData      = "a60ae97c-268c-40cb-af5f-dd276e917ed7"
	osID             = "7d1914e1-f4ab-47fc-bd8c-b3a23143e87a"
	incompatibleOsIS = "4aff4d9d-b1f4-44b0-ab6f-e4711ac11711"
)

func TestAccDataSourceEasyParitioning_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
data "scaleway_easy_partitioning" "test" {
	offer_id = "%s"
	os_id = "%s"
}
`, offerISData, osID,
				),
			},
		},
	})
}
