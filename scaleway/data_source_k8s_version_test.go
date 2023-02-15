package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
)

func TestAccScalewayDataSourceK8SVersion_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_k8s_version" "by_name" {
						name = "1.26.0"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SVersionExists(tt, "data.scaleway_k8s_version.by_name"),
					resource.TestCheckResourceAttrSet("data.scaleway_k8s_version.by_name", "name"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.#", "2"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.0", "cilium"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.1", "calico"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_container_runtimes.#", "1"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_container_runtimes.0", "containerd"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.#", "3"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.0", "HPAScaleToZero"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.1", "GRPCContainerProbe"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.2", "ReadWriteOncePod"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SVersionExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		region, name, err := parseRegionalID(rs.Primary.ID)
		if err != nil {
			return err
		}

		k8sAPI := k8s.NewAPI(tt.Meta.scwClient)
		_, err = k8sAPI.GetVersion(&k8s.GetVersionRequest{
			Region:      region,
			VersionName: name,
		})

		if err != nil {
			return err
		}

		return nil
	}
}
