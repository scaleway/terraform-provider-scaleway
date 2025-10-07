package instance_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	product_catalog "github.com/scaleway/scaleway-sdk-go/api/product_catalog/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestAccDataSourceServerType_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_instance_server_type" "dev" {
						name = "DEV1-XL"
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "name", "DEV1-XL"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "arch", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "cpu", "4"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "ram", "12884901888"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "gpu", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.min_size_total", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.max_size_total", "120000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.min_size_per_local_volume", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.max_size_per_local_volume", "800000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.scratch_storage_max_size", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "volumes.0.block_storage", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.internal_bandwidth", "500000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.public_bandwidth", "500000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.block_bandwidth", "262144000"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.dev", "hourly_price"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "end_of_service", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.dev", "availability"),
				),
			},
			{
				Config: `
					data "scaleway_instance_server_type" "gpu" {
						name = "RENDER-S"
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "name", "RENDER-S"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "arch", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "cpu", "10"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "ram", "45097156608"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "gpu", "1"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.min_size_total", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.max_size_total", "400000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.min_size_per_local_volume", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.max_size_per_local_volume", "800000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.scratch_storage_max_size", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "volumes.0.block_storage", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.internal_bandwidth", "2000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.public_bandwidth", "2000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.block_bandwidth", "2147483648"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.gpu", "hourly_price"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "end_of_service", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.gpu", "availability"),
				),
			},
			{
				Config: `
					data "scaleway_instance_server_type" "pro" {
						name = "PRO2-XXS"
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "name", "PRO2-XXS"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "gpu", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "cpu", "2"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "arch", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "ram", "8589934592"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.min_size_total", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.max_size_total", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.min_size_per_local_volume", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.max_size_per_local_volume", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.scratch_storage_max_size", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "volumes.0.block_storage", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.internal_bandwidth", "350000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.public_bandwidth", "350000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.block_bandwidth", "131072000"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.pro", "hourly_price"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "end_of_service", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_server_type.pro", "availability"),
				),
			},
		},
	})
}

func TestAccDataSourceServerType_CompareWithPCU(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	client := meta.ExtractScwClient(tt.Meta)
	pcuAPI := product_catalog.NewPublicCatalogAPI(client)
	steps := []resource.TestStep(nil)

	serverTypeToTestByZone := map[scw.Zone]string{
		scw.ZoneFrPar1: "RENDER-S",
		scw.ZoneFrPar2: "H100-SXM-2-80G",
		scw.ZoneFrPar3: "POP2-HM-16C-128G",
		scw.ZoneNlAms1: "COPARM1-8C-32G",
		scw.ZoneNlAms2: "DEV1-XL",
		scw.ZoneNlAms3: "PLAY2-NANO",
		scw.ZonePlWaw1: "GP1-L",
		scw.ZonePlWaw2: "L4-1-24G",
		scw.ZonePlWaw3: "PRO2-M",
	}

	for zone, serverTypeToTest := range serverTypeToTestByZone {
		// List all available server types in the zone to test
		pcuInstances, err := pcuAPI.ListPublicCatalogProducts(&product_catalog.PublicCatalogAPIListPublicCatalogProductsRequest{
			ProductTypes: []product_catalog.ListPublicCatalogProductsRequestProductType{
				product_catalog.ListPublicCatalogProductsRequestProductTypeInstance,
			},
			Zone: &zone,
		}, scw.WithAllPages(), scw.WithContext(t.Context()))
		if err != nil {
			t.Fatal(err)
		}

		// Look for the server type to test in the PCU
		for _, pcuInstance := range pcuInstances.Products {
			if pcuInstance.Properties.Instance.OfferID != serverTypeToTest {
				continue
			}

			// Fetch expected values from the PCU to be compared with the data source's info
			datasourceTFName := "data.scaleway_instance_server_type." + serverTypeToTest
			hardwareSpecs := pcuInstance.Properties.Hardware

			expectedArch := ""

			switch hardwareSpecs.CPU.Arch {
			case product_catalog.PublicCatalogProductPropertiesHardwareCPUArchX64:
				expectedArch = instance.ArchX86_64.String()
			case product_catalog.PublicCatalogProductPropertiesHardwareCPUArchArm64:
				expectedArch = instance.ArchArm64.String()
			case product_catalog.PublicCatalogProductPropertiesHardwareCPUArchUnknownArch:
				expectedArch = instance.ArchUnknownArch.String()
			}

			expectedCPU := strconv.FormatUint(uint64(hardwareSpecs.CPU.Threads), 10)
			expectedRAM := hardwareSpecs.RAM.Size.String()

			expectedGPU := "0"
			if hardwareSpecs.Gpu != nil {
				expectedGPU = strconv.FormatUint(uint64(hardwareSpecs.Gpu.Count), 10)
			}

			expectedInternalBandwidth := strconv.FormatUint(hardwareSpecs.Network.InternalBandwidth, 10)
			expectedPublicBandwidth := strconv.FormatUint(hardwareSpecs.Network.PublicBandwidth, 10)

			expectedHourlyPrice := strings.TrimPrefix(pcuInstance.Price.RetailPrice.String(), "€ ")

			// Create test step
			steps = append(steps, resource.TestStep{
				Config: fmt.Sprintf(`
				data "scaleway_instance_server_type" "%[1]s" {
					name = "%[1]s"
					zone = "%[2]s"
				}`, serverTypeToTest, zone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceTFName, "name", serverTypeToTest),
					resource.TestCheckResourceAttr(datasourceTFName, "zone", zone.String()),
					resource.TestCheckResourceAttr(datasourceTFName, "arch", expectedArch),
					resource.TestCheckResourceAttr(datasourceTFName, "cpu", expectedCPU),
					resource.TestCheckResourceAttr(datasourceTFName, "ram", expectedRAM),
					resource.TestCheckResourceAttr(datasourceTFName, "gpu", expectedGPU),
					resource.TestCheckResourceAttr(datasourceTFName, "network.0.internal_bandwidth", expectedInternalBandwidth),
					resource.TestCheckResourceAttr(datasourceTFName, "network.0.public_bandwidth", expectedPublicBandwidth),
					resource.TestCheckResourceAttr(datasourceTFName, "hourly_price", expectedHourlyPrice),
				),
			})
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps:             steps,
	})
}
