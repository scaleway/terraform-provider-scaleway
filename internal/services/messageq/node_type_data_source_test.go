package messageq_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceMessageQNodeType_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "scaleway_messageq_node_type" "main" {
  name = "%s"
}
`, nodeType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_messageq_node_type.main", "name", nodeType),
					resource.TestCheckResourceAttrSet("data.scaleway_messageq_node_type.main", "vcpus"),
					resource.TestCheckResourceAttrSet("data.scaleway_messageq_node_type.main", "memory_size_in_gb"),
					resource.TestCheckResourceAttrSet("data.scaleway_messageq_node_type.main", "stock_status"),
				),
			},
		},
	})
}
