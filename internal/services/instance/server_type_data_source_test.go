package instance_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "hourly_price", "0.07308000326156616"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.internal_bandwidth", "500000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.public_bandwidth", "500000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.block_bandwidth", "262144000"),
					// resource.TestCheckResourceAttr("data.scaleway_instance_server_type.dev", "network.0.ipv6_support", "true"),
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
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "hourly_price", "1.2425999641418457"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.internal_bandwidth", "2000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.public_bandwidth", "2000000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.block_bandwidth", "2147483648"),
					// resource.TestCheckResourceAttr("data.scaleway_instance_server_type.gpu", "network.0.ipv6_support", "true"),
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
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "hourly_price", "0.054999999701976776"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.boot_types.0", "local"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.boot_types.1", "rescue"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "capabilities.0.max_file_systems", "0"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.internal_bandwidth", "350000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.public_bandwidth", "350000000"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.block_bandwidth", "131072000"),
					// resource.TestCheckResourceAttr("data.scaleway_instance_server_type.pro", "network.0.ipv6_support", "true"),
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

	serversTypesToTestByZone := map[scw.Zone][]string{
		scw.ZoneFrPar1: {
			"COPARM1-2C-8G",
			"DEV1-M",
			"RENDER-S",
			"STARDUST1-S",
		},
		scw.ZoneFrPar2: {
			"COPARM1-32C-128G",
			"DEV1-L",
			"GP1-M",
			"H100-SXM-2-80G",
			"PLAY2-MICRO",
		},
		scw.ZoneFrPar3: {
			"GP1-S",
			"POP2-HM-16C-128G",
			"PRO2-XXS",
		},
		scw.ZoneNlAms1: {
			"COPARM1-8C-32G",
			"GP1-XL",
			"PRO2-XS",
		},
		scw.ZoneNlAms2: {
			"DEV1-XL",
			"POP2-HN-5",
			"PRO2-S",
		},
		scw.ZoneNlAms3: {
			"PLAY2-NANO",
			"POP2-2C-8G",
			"PRO2-L",
		},
		scw.ZonePlWaw1: {
			"DEV1-S",
			"GP1-L",
			"POP2-64C-256G",
		},
		scw.ZonePlWaw2: {
			"GP1-XS",
			"H100-1-80G",
			"L4-1-24G",
			"L40S-4-48G",
		},
		scw.ZonePlWaw3: {
			"PLAY2-PICO",
			"POP2-HC-32C-64G",
			"PRO2-M",
		},
	}

	for zone, serverTypesToTest := range serversTypesToTestByZone {
		// List all available server types in selected zone
		pcuInstances, err := pcuAPI.ListPublicCatalogProducts(&product_catalog.PublicCatalogAPIListPublicCatalogProductsRequest{
			ProductTypes: []product_catalog.ListPublicCatalogProductsRequestProductType{
				product_catalog.ListPublicCatalogProductsRequestProductTypeInstance,
			},
			Zone: &zone,
			// TODO: Global does not work
			// Global: scw.BoolPtr(true),
		}, scw.WithAllPages(), scw.WithContext(t.Context()))
		if err != nil {
			t.Fatal(err)
		}

		// Look for each server type to be tested in the zone in the PCU
		for _, serverTypeToTest := range serverTypesToTest {
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
				// TODO: prices differ between Instance and PCU
				// expectedHourlyPrice := strconv.FormatFloat(serverTypeToTest.Price.RetailPrice.ToFloat(), 'f', -1, 64)

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
						// resource.TestCheckResourceAttr(datasourceTFName, "hourly_price", expectedHourlyPrice),
					),
				})
			}
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps:             steps,
	})
}
