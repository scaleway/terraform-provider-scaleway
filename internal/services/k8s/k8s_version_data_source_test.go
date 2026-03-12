package k8s_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	k8sSDK "github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccDataSourceVersion_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	version := "1.32.2"
	versionWithoutPatch := "1.32"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_k8s_version" "by_name" {
						name = %q
					}
				`, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckK8SVersionExists(tt, "data.scaleway_k8s_version.by_name"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "name", version),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "major_minor_only", versionWithoutPatch),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.#", "4"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.0", "cilium"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.1", "calico"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.2", "kilo"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_cnis.3", "none"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_container_runtimes.#", "1"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_container_runtimes.0", "containerd"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.#", "9"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.0", "HPAScaleToZero"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.1", "InPlacePodVerticalScaling"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.2", "SidecarContainers"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.3", "DRAAdminAccess"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.4", "DRAResourceClaimDeviceStatus"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.5", "DynamicResourceAllocation"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.6", "PodLevelResources"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.7", "CPUManagerPolicyAlphaOptions"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.by_name", "available_feature_gates.8", "ImageVolume"),
				),
			},
		},
	})
}

func TestAccDataSourceVersion_Latest(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_k8s_version" "latest" {
						name = "latest"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckK8SVersionExists(tt, "data.scaleway_k8s_version.latest"),
					resource.TestCheckResourceAttrSet("data.scaleway_k8s_version.latest", "name"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.latest", "name", testAccK8SClusterGetLatestK8SVersion(tt)),
				),
			},
		},
	})
}

func TestAccDataSourceVersion_WithAutoUpgrade(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccK8SClusterGetLatestK8SVersion(tt)

	latestK8SVersionWithoutPatch, err := k8s.VersionNameWithoutPatch(latestK8SVersion)
	if err != nil {
		t.Fatal(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_k8s_version" "latest" {
						name = "latest"
					}
					` + testAccCheckK8SClusterAutoUpgrade(true, "any", 0, "data.scaleway_k8s_version.latest.major_minor_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckK8SVersionExists(tt, "data.scaleway_k8s_version.latest"),
					testAccCheckK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.latest", "name", latestK8SVersion),
					resource.TestCheckResourceAttr("data.scaleway_k8s_version.latest", "major_minor_only", latestK8SVersionWithoutPatch),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", latestK8SVersionWithoutPatch),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
		},
	})
}

func testAccCheckK8SVersionExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		region, name, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		k8sAPI := k8sSDK.NewAPI(tt.Meta.ScwClient())

		_, err = k8sAPI.GetVersion(&k8sSDK.GetVersionRequest{
			Region:      region,
			VersionName: name,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func TestVersionNameWithoutPatch(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	t.Run("ok-without-prefix", func(t *testing.T) {
		version := "1.32.3"
		expected := "1.32"
		actual, err := k8s.VersionNameWithoutPatch(version)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("ok-with-prefix", func(t *testing.T) {
		version := "v2.57.9"
		expected := "v2.57"
		actual, err := k8s.VersionNameWithoutPatch(version)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("errors", func(t *testing.T) {
		versionsToTest := []string{
			"1.32.3.4",
			"1.32",
			"",
		}
		for _, version := range versionsToTest {
			expectedError := "version name must contain 3 parts"
			_, err := k8s.VersionNameWithoutPatch(version)
			assert.ErrorContains(t, err, expectedError)
		}
	})
}
